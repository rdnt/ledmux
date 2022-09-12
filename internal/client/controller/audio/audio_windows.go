package audio

import (
	"context"
	"fmt"
	"math"
	"math/cmplx"
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/pkg/errors"
	"github.com/sgreben/piecewiselinear"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/dsp/window"

	"ledctl3/internal/client/visualizer"
	"ledctl3/pkg/gradient"
)

type Visualizer struct {
	mux sync.Mutex

	leds     int
	colors   []colorful.Color
	segments []Segment

	events      chan visualizer.UpdateEvent
	cancel      context.CancelFunc
	childCancel context.CancelFunc
	done        chan bool
	maxLedCount int

	processing bool

	gradient   gradient.Gradient
	windowSize int
}

func (v *Visualizer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	v.done = make(chan bool)

	go func() {
		for {
			select {
			case <-ctx.Done():
				v.done <- true
				return
			default:
				var childCtx context.Context
				childCtx, v.childCancel = context.WithCancel(ctx)

				err := v.startCapture(childCtx)
				if errors.Is(err, context.Canceled) {
					return
				} else if err != nil {
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()

	return nil
}

func (v *Visualizer) Events() chan visualizer.UpdateEvent {
	return v.events
}

func (v *Visualizer) Stop() error {
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}

	<-v.done

	return nil
}

func (v *Visualizer) startCapture(ctx context.Context) error {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	if err := mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return err
	}
	defer mmd.Release()

	var ps *wca.IPropertyStore
	if err := mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		return err
	}
	defer ps.Release()

	var ac *wca.IAudioClient
	if err := mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &ac); err != nil {
		return err
	}
	defer ac.Release()

	var wfx *wca.WAVEFORMATEX
	if err := ac.GetMixFormat(&wfx); err != nil {
		return err
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	wfx.NChannels = 2 // force channels to two
	wfx.WFormatTag = 1
	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
	wfx.CbSize = 0

	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	fmt.Printf("Channels: %d\n", wfx.NChannels)

	var defaultPeriod wca.REFERENCE_TIME
	var minimumPeriod wca.REFERENCE_TIME
	var latency time.Duration
	if err := ac.GetDevicePeriod(&defaultPeriod, &minimumPeriod); err != nil {
		return err
	}
	latency = time.Duration(int(minimumPeriod) * 100)

	fmt.Println("Default period: ", defaultPeriod)
	fmt.Println("Minimum period: ", minimumPeriod)
	fmt.Println("Latency: ", latency)

	if err := ac.Initialize(
		wca.AUDCLNT_SHAREMODE_SHARED,
		wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK|wca.AUDCLNT_STREAMFLAGS_LOOPBACK,
		defaultPeriod, 0, wfx, nil,
	); err != nil {
		panic(err)
		return err
	}

	audioReadyEvent := wca.CreateEventExA(0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE)
	defer wca.CloseHandle(audioReadyEvent)

	if err := ac.SetEventHandle(audioReadyEvent); err != nil {
		return err
	}

	var bufferFrameSize uint32
	if err := ac.GetBufferSize(&bufferFrameSize); err != nil {
		return err
	}
	fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)

	var acc *wca.IAudioCaptureClient
	if err := ac.GetService(wca.IID_IAudioCaptureClient, &acc); err != nil {
		return err
	}
	defer acc.Release()

	if err := ac.Start(); err != nil {
		return err
	}

	var offset int
	var b *byte
	var data *byte
	var availableFrameSize uint32
	var flags uint32
	var devicePosition uint64
	var qcpPosition uint64

	errorChan := make(chan error, 1)

	var isCapturing = true

loop:
	for {
		if !isCapturing {
			close(errorChan)
			break
		}
		go func() {
			errorChan <- watchEvent(ctx, audioReadyEvent)
		}()

		select {
		case <-ctx.Done():
			isCapturing = false
			<-errorChan
			break loop
		case err := <-errorChan:
			if err != nil {
				isCapturing = false
				break
			}

			if err = acc.GetBuffer(&data, &availableFrameSize, &flags, &devicePosition, &qcpPosition); err != nil {
				continue
			}

			if availableFrameSize == 0 {
				continue
			}

			start := unsafe.Pointer(data)
			if start == nil {
				continue
			}

			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
			buf := make([]byte, lim)

			for n := 0; n < lim; n++ {
				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
				buf[n] = *b
			}

			offset += lim

			samples := make([]float64, len(buf)/4)
			for i := 0; i < len(buf); i += 4 {
				v := float64(readInt32(buf[i : i+4]))
				samples = append(samples, v)
			}

			go v.process(samples)

			if err = acc.ReleaseBuffer(availableFrameSize); err != nil {
				return errors.WithMessage(err, "failed to ReleaseBuffer")
			}
		}

	}

	if err := ac.Stop(); err != nil {
		return errors.Wrap(err, "failed to stop audio client")
	}

	return nil
}

func watchEvent(ctx context.Context, event uintptr) (err error) {
	errorChan := make(chan error, 1)
	go func() {
		errorChan <- eventEmitter(event)
	}()
	select {
	case err = <-errorChan:
		close(errorChan)
		return
	case <-ctx.Done():
		err = ctx.Err()
		return
	}
}

func eventEmitter(event uintptr) (err error) {
	// if err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
	//	return
	// }
	dw := wca.WaitForSingleObject(event, wca.INFINITE)
	if dw != 0 {
		return fmt.Errorf("failed to watch event")
	}
	// ole.CoUninitialize()
	return
}

// readInt32 reads a signed integer from a byte slice. only a slice with len(4)
// should be passed. equivalent of int32(binary.LittleEndian.Uint32(b))
func readInt32(b []byte) int32 {
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
}

func SpanLog(min, max float64, nPoints int) []float64 {
	X := make([]float64, nPoints)
	min, max = math.Min(max, min), math.Max(max, min)
	d := max - min
	for i := range X {
		v := min + d*(float64(i)/float64(nPoints-1))
		v = math.Pow(v, 0.5)
		X[i] = v
	}
	return X
}

func (v *Visualizer) process(samples []float64) {
	now := time.Now()

	v.mux.Lock()
	if v.processing {
		v.mux.Unlock()
		return
	}

	v.processing = true
	v.mux.Unlock()

	defer func() {
		v.mux.Lock()
		v.processing = false
		v.mux.Unlock()
	}()

	fft := fourier.NewFFT(len(samples))
	coeff := fft.Coefficients(nil, window.Hamming(samples))

	freqs := []float64{}
	var maxfreq float64
	for _, c := range coeff {
		freqs = append(freqs, cmplx.Abs(c))
		if cmplx.Abs(c) > maxfreq {
			maxfreq = cmplx.Abs(c)
		}
	}

	for i, f := range freqs {
		norm := normalize(float64(int(f)), 0, maxfreq)
		freqs[i] = norm
	}

	// Only keep the first half of the fft
	freqs = freqs[:len(freqs)/2]

	// Scale the frequencies so that low ones are more pronounced.
	f := piecewiselinear.Function{Y: freqs}
	f.X = SpanLog(0, 1, len(f.Y))

	freqs = make([]float64, v.maxLedCount)
	for i := 0; i < v.maxLedCount; i++ {
		freqs[i] = f.At(float64(i) / float64(v.maxLedCount-1))
	}

	pix := []byte{}

	maxLeds := v.maxLedCount

	for i := 0; i < maxLeds; i++ {
		freq := freqs[i]

		c := v.gradient.GetInterpolatedColor(freq)

		hue, sat, val := c.Hsl()

		val = math.Sqrt(1 - math.Pow(freq-1, 2))
		val = math.Min(val, 1)
		val = math.Max(val, 0.25)

		c = colorful.Hsv(hue, sat, val)

		r, g, b, _ := c.RGBA()

		r = r >> 8
		g = g >> 8
		b = b >> 8

		pix = append(pix, []byte{uint8(r), uint8(g), uint8(b), 0xFF}...)
	}

	pixs = append(pixs, pix)
	if len(pixs) > v.windowSize {
		pixs = pixs[1:]
	}

	weights := []float64{}
	weightsTotal := 0.0

	for i := 0; i < len(pixs); i++ {
		// for each history item
		w := float64((i+1)*(i+1) + len(pixs))

		weights = append(weights, w)
		weightsTotal += w
	}

	pix2 := make([]float64, len(pix))
	for i, p2 := range pixs {
		for j, p := range p2 {
			pix2[j] = pix2[j] + float64(p)*weights[i]
		}
	}

	pix3 := make([]float64, len(pix))
	for i, p := range pix2 {
		avg := p / weightsTotal
		pix3[i] = float64(avg)
	}

	segs := []visualizer.Segment{}

	for _, seg := range v.segments {
		length := seg.Leds * 4
		pix4 := make([]uint8, length)

		for i := 0; i < length; i += 4 {
			offset := i

			// TODO: do the mirroring beforehand (not with the 2nd part of the fft...)
			//  by limiting max to maxleds/2 and then flipping the first half into the second
			if i >= length/2 {
				offset = length - 4 - i
			}

			pix4[i] = uint8(pix3[offset])
			pix4[i+1] = uint8(pix3[offset+1])
			pix4[i+2] = uint8(pix3[offset+2])
			pix4[i+3] = uint8(pix3[offset+3])
		}

		pix := pix4[:seg.Leds*4]

		// if seg.Id == 0 {
		// 	out := "\r"
		// 	//out := "\n"
		// 	for i := 0; i < len(pix); i += 4 {
		// 		out += color.RGB(pix[i], pix[i+1], pix[i+2], true).Sprintf(" ")
		// 	}
		// 	fmt.Print(out)
		// }

		segs = append(segs, visualizer.Segment{
			Id:  seg.Id,
			Pix: pix,
		})
	}

	v.events <- visualizer.UpdateEvent{
		Segments: segs,
		Duration: time.Since(now),
	}
}

// normalize scales a value from min,max to 0,1
func normalize(val, min, max float64) float64 {
	if max == min {
		return max
	}

	return (val - min) / (max - min)
}

var pixs [][]byte

func New(opts ...Option) (v *Visualizer, err error) {
	v = new(Visualizer)

	for _, opt := range opts {
		err := opt(v)
		if err != nil {
			return nil, err
		}
	}

	v.gradient, err = gradient.New(v.colors...)
	if err != nil {
		return nil, err
	}

	v.events = make(chan visualizer.UpdateEvent, len(v.segments)*8)

	return v, nil
}

type Segment struct {
	Id   int
	Leds int
}
