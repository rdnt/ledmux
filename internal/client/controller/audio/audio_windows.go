package audio

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/eripe970/go-dsp-utils"
	"github.com/go-ole/go-ole"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/pkg/errors"

	"ledctl3/internal/client/visualizer"
)

type Visualizer struct {
	mux         sync.Mutex
	leds        int
	events      chan visualizer.UpdateEvent
	cancel      context.CancelFunc
	childCancel context.CancelFunc
	done        chan bool
	segments    []Segment
	maxLedCount int

	processing bool

	config Config
}

type Config struct {
	Color1 colorful.Color
	Color2 colorful.Color
	Color3 colorful.Color
	Color4 colorful.Color
	Color5 colorful.Color
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

			go v.processBuf(buf, float64(wfx.NSamplesPerSec))

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
	return
}

func eventEmitter(event uintptr) (err error) {
	//if err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
	//	return
	//}
	dw := wca.WaitForSingleObject(event, wca.INFINITE)
	if dw != 0 {
		return fmt.Errorf("failed to watch event")
	}
	//ole.CoUninitialize()
	return
}

func (v *Visualizer) processBuf(buf []byte, samplesPerSec float64) {
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

	samples := make([]float64, len(buf)/4)
	for i := 0; i < len(buf); i += 4 {
		v := float64(int32(binary.LittleEndian.Uint32(buf[i : i+4])))

		samples = append(samples, v)
	}

	signal := dsp.Signal{
		SampleRate: samplesPerSec,
		Signal:     samples,
	}

	var freqs []float64

	normalized, err := signal.Normalize()
	if err != nil {
		freqs = make([]float64, 882)
	} else {
		spectrum, _ := normalized.FrequencySpectrum() // never fails
		// freqs length will be half of the samples length
		freqs = spectrum.Spectrum
	}

	lows := []float64{}

	// make the low frequencies more prominent
	lows = append(lows, freqs[0])
	lows = append(lows, freqs[1])
	lows = append(lows, freqs[1])
	lows = append(lows, freqs[1])
	lows = append(lows, freqs[2])
	lows = append(lows, freqs[2])
	lows = append(lows, freqs[2])
	lows = append(lows, freqs[3])
	lows = append(lows, freqs[3])
	lows = append(lows, freqs[3])
	lows = append(lows, freqs[4])
	lows = append(lows, freqs[4])
	lows = append(lows, freqs[5])
	lows = append(lows, freqs[5])

	freqs = append(lows, freqs[6:]...)

	max := 0.0
	for _, freq := range freqs {
		if math.Abs(freq) > max {
			max = math.Abs(freq)
		}
	}

	pix := []byte{}

	// TODO: min func for ints
	for i := 0; i < int(math.Min(float64(v.maxLedCount), float64(len(freqs)))); i++ {
		var curr float64
		// diffuse frequencies horizontally (just a little)
		if i < len(freqs)-3 {
			curr = freqs[i] + freqs[i+1]
		} else {
			curr = freqs[i] + freqs[i]
		}
		curr = (curr) / 2

		norm := normalize(curr, 0, max)

		var c colorful.Color
		if norm < 0.3 {
			c = v.config.Color1
		} else if norm < 0.45 {
			c = v.config.Color2
		} else if norm < 0.8 {
			c = v.config.Color3
		} else if norm < 0.9 {
			c = v.config.Color4
		} else {
			c = v.config.Color5
		}

		h, s, v := c.Hsv()

		s = math.Min(max, 1)*0.25 + s*0.75
		v = math.Min(v*1.1, 1)

		c = colorful.Hsv(h, s, v)

		r, g, b, _ := c.RGBA()

		r = uint32(math.Min(float64(r/256), 255))
		g = uint32(math.Min(float64(g/256), 255))
		b = uint32(math.Min(float64(b/256), 255))

		pix = append(pix, []byte{uint8(r), uint8(g), uint8(b), 0xFF}...)
	}

	pixs = append(pixs, pix)
	if len(pixs) > 100 {
		pixs = pixs[1:]
	}

	weights := []float64{}
	weightsTotal := 0.0

	for i := 0; i < len(pixs); i++ {
		w := float64(i + len(pixs))
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
		// 	out := "\n"
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

func normalize(val, min, max float64) float64 {
	if max == min {
		return max
	}

	return (val - min) / (max - min)
}

var pixs [][]byte

func New(opts Options) (*Visualizer, error) {
	v := &Visualizer{}

	for _, seg := range opts.Segments {
		if seg.Leds > v.maxLedCount {
			v.maxLedCount = seg.Leds
		}
	}

	v.segments = opts.Segments
	v.leds = opts.Leds
	v.events = make(chan visualizer.UpdateEvent, len(v.segments)*8)

	c1, _ := colorful.Hex("#110022")
	c2, _ := colorful.Hex("#602980")
	c3, _ := colorful.Hex("#442968")
	c4, _ := colorful.Hex("#2ffee1")
	c5, _ := colorful.Hex("#ea267a")

	v.config = Config{
		Color1: c1,
		Color2: c2,
		Color3: c3,
		Color4: c4,
		Color5: c5,
	}

	return v, nil
}

type Segment struct {
	Id   int
	Leds int
}
