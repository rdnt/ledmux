package audioviz

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"syscall"
	"time"
	"unsafe"

	"github.com/eripe970/go-dsp-utils"
	"github.com/go-ole/go-ole"
	"github.com/gookit/color"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/pkg/errors"

	"ledctl3/internal/client/interfaces"
)

type IAudioMeterInformation struct {
	ole.IUnknown
}

type IAudioMeterInformationVtbl struct {
	ole.IUnknownVtbl
	GetPeakValue            uintptr
	GetChannelsPeakValues   uintptr
	GetMeteringChannelCount uintptr
	QueryHardwareSupport    uintptr
}

func (v *IAudioMeterInformation) VTable() *IAudioMeterInformationVtbl {
	return (*IAudioMeterInformationVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *IAudioMeterInformation) GetPeakValue(peak *float32) (err error) {
	err = amiGetPeakValue(v, peak)
	return
}

func (v *IAudioMeterInformation) GetMeteringChannelCount(count *uint32) (err error) {
	err = amiGetMeteringChannelCount(v, count)
	return
}

func (v *IAudioMeterInformation) GetChannelsPeakValues(count uint32, peaks []float32) (err error) {
	err = amiGetChannelsPeakValues(v, count, peaks)
	return
}

func (v *IAudioMeterInformation) QueryHardwareSupport(response *uint32) (err error) {
	err = amiQueryHardwareSupport(v, response)
	return
}

func amiGetPeakValue(ami *IAudioMeterInformation, peak *float32) (err error) {
	hr, _, _ := syscall.Syscall(
		ami.VTable().GetPeakValue,
		2,
		uintptr(unsafe.Pointer(ami)),
		uintptr(unsafe.Pointer(peak)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return

}

func amiGetChannelsPeakValues(ami *IAudioMeterInformation, count uint32, peaks []float32) (err error) {
	hr, _, _ := syscall.Syscall(ami.VTable().GetChannelsPeakValues,
		3,
		uintptr(unsafe.Pointer(ami)),
		uintptr(count),
		uintptr(unsafe.Pointer(&peaks[0])))
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func amiGetMeteringChannelCount(ami *IAudioMeterInformation, count *uint32) (err error) {
	hr, _, _ := syscall.Syscall(
		ami.VTable().GetMeteringChannelCount,
		2,
		uintptr(unsafe.Pointer(ami)),
		uintptr(unsafe.Pointer(count)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

func amiQueryHardwareSupport(ami *IAudioMeterInformation, response *uint32) (err error) {
	hr, _, _ := syscall.Syscall(
		ami.VTable().GetMeteringChannelCount,
		2,
		uintptr(unsafe.Pointer(ami)),
		uintptr(unsafe.Pointer(response)),
		0)
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}

type Visualizer struct {
	source         interfaces.AudioSource
	leds           int
	events         chan interfaces.UpdateEvent
	cancel         context.CancelFunc
	childCancel    context.CancelFunc
	done           chan bool
	lastSourceName string
	segments       []Segment
	maxLedCount    int
}

func (v *Visualizer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	v.done = make(chan bool)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("parent ctx done, exiting")
				v.done <- true
				return
			default:
				fmt.Println("STARTING AUDIO CAPTURE")

				var childCtx context.Context
				childCtx, v.childCancel = context.WithCancel(ctx)

				err := v.startCapture(childCtx)
				if errors.Is(err, context.Canceled) {
					fmt.Println("capture canceled")

					return
				} else if err != nil {
					fmt.Println("error starting capture:", err)

					time.Sleep(1 * time.Second)
				}
			}
		}
	}()

	return nil
}

func (v *Visualizer) Events() chan interfaces.UpdateEvent {
	return v.events
}

func (v *Visualizer) Stop() error {
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}

	fmt.Println("stop: waiting for done")

	// FIXME: this blocks shutdown for some reason..
	//<-v.done

	fmt.Println("stop: done received")
	return nil
}

//func (v *Visualizer) readStateChanges() error {
//
//	time.Sleep(10 * time.Minute)
//
//	fmt.Println("Done")
//
//	return nil
//}

var v2 *Visualizer

func onDefaultDeviceChanged(flow wca.EDataFlow, role wca.ERole, pwstrDeviceId string) error {
	fmt.Printf("Called OnDefaultDeviceChanged\t(%v, %v, %q)\n", flow, role, pwstrDeviceId)
	//time.Sleep(1 * time.Second)

	v2.childCancel()

	time.Sleep(3 * time.Second)
	fmt.Println("restart")

	//err := v2.Stop()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//err = v2.Start()
	//if err != nil {
	//	fmt.Println(err)
	//}
	return nil
}

func onDeviceAdded(pwstrDeviceId string) error {
	fmt.Printf("Called OnDeviceAdded\t(%q)\n", pwstrDeviceId)

	return nil
}

func onDeviceRemoved(pwstrDeviceId string) error {
	fmt.Printf("Called OnDeviceRemoved\t(%q)\n", pwstrDeviceId)

	return nil
}

func onDeviceStateChanged(pwstrDeviceId string, dwNewState uint64) error {
	fmt.Printf("Called OnDeviceStateChanged\t(%q, %v)\n", pwstrDeviceId, dwNewState)

	return nil
}

func onPropertyValueChanged(pwstrDeviceId string, key uint64) error {
	fmt.Printf("Called OnPropertyValueChanged\t(%q, %v)\n", pwstrDeviceId, key)
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

	v2 = v

	//callback := wca.IMMNotificationClientCallback{
	//	//OnDefaultDeviceChanged: onDefaultDeviceChanged,
	//	//OnDeviceAdded:          onDeviceAdded,
	//	//OnDeviceRemoved:        onDeviceRemoved,
	//	//OnDeviceStateChanged:   onDeviceStateChanged,
	//	//OnPropertyValueChanged: onPropertyValueChanged,
	//}
	//
	//mmnc := wca.NewIMMNotificationClient(callback)
	//
	//if err := mmde.RegisterEndpointNotificationCallback(mmnc); err != nil {
	//	return errors.WithMessage(err, "failed to RegisterEndpointNotificationCallback readstate")
	//}

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

	var pv wca.PROPVARIANT
	if err := ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		return err
	}
	fmt.Printf("Capturing audio from: %s\n", pv.String())

	var ac *wca.IAudioClient
	if err := mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &ac); err != nil {
		return err
	}
	defer ac.Release()

	var vol *wca.IAudioEndpointVolume
	if err := mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &vol); err != nil {
		return err
	}
	defer vol.Release()

	var ami *IAudioMeterInformation
	if err := mmd.Activate(wca.IID_IAudioMeterInformation, wca.CLSCTX_ALL, nil, &ami); err != nil {
		return err
	}
	defer ami.Release()

	var fPeak float32
	if err := ami.GetPeakValue(&fPeak); err != nil {
		return err
	}
	fmt.Printf("Peak: %f\n", fPeak)

	//go func() {
	//	for {
	//		time.Sleep(1 * time.Second)
	//		if err := ami.GetPeakValue(&fPeak); err != nil {
	//			fmt.Println(err)
	//		}
	//
	//		fmt.Printf("Peak: %f\n", fPeak)
	//	}
	//}()

	var fVolume float32
	if err := vol.GetMasterVolumeLevelScalar(&fVolume); err != nil {
		return err
	}
	fmt.Printf("Volume: %f\n", fVolume)

	var chanVol float32
	if err := vol.GetChannelVolumeLevelScalar(1, &fVolume); err != nil {
		return err
	}
	fmt.Printf("Channel scalar volume: %f\n", chanVol)

	var wfx *wca.WAVEFORMATEX
	if err := ac.GetMixFormat(&wfx); err != nil {
		return err
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	wfx.WFormatTag = 1
	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
	wfx.CbSize = 0

	fmt.Println("--------")
	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	fmt.Printf("Channels: %d\n", wfx.NChannels)
	fmt.Println("--------")

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

	if err := ac.Initialize(wca.AUDCLNT_SHAREMODE_SHARED, wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK|wca.AUDCLNT_STREAMFLAGS_LOOPBACK, defaultPeriod, 0, wfx, nil); err != nil {
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
	fmt.Println("Start capturing with shared event driven mode")

	var offset int
	//var isCapturing = true
	var b *byte
	var data *byte
	var availableFrameSize uint32
	var flags uint32
	var devicePosition uint64
	var qcpPosition uint64

	errorChan := make(chan error, 1)

	//time.Sleep(latency)

	//var padding uint32

	//in := make(chan float32)
	//out := make(chan float32)

	//go bpm.ProgressivelyReadFloatArray(in, out)

	//done := make(chan bool)

	//go readProgressiveVars(out, done, *progressive, *progressiveInterval)

	//var scores []float64

	var isCapturing = true

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
			break
		case err := <-errorChan:
			if err != nil {
				isCapturing = false
				break
			}

			if err := ami.GetPeakValue(&fPeak); err != nil {
				return errors.WithMessage(err, "failed to get peak")
			}

			if err = acc.GetBuffer(&data, &availableFrameSize, &flags, &devicePosition, &qcpPosition); err != nil {
				fmt.Println("GetBuffer failed", err)
				continue
			}

			if availableFrameSize == 0 {
				fmt.Println("0 frame size")
				continue
			}

			start := unsafe.Pointer(data)
			if start == nil {
				fmt.Println("nil data")
				continue
			}

			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
			buf := make([]byte, lim)

			for n := 0; n < lim; n++ {
				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
				buf[n] = *b
			}

			offset += lim

			//if offset%(lim*4) == 0 {
			//	go v.processBuf(buf, float64(fPeak))
			//}

			go v.processBuf(buf, float64(fPeak), float64(wfx.NSamplesPerSec))

			if err = acc.ReleaseBuffer(availableFrameSize); err != nil {
				return errors.WithMessage(err, "failed to ReleaseBuffer")
			}
		}

	}

	//for {
	//
	//	if !isCapturing {
	//		break
	//	}
	//	select {
	//	case <-ctx.Done():
	//		isCapturing = false
	//		break
	//	default:
	//		// Wait for buffering.
	//		//time.Sleep(latency / 2)
	//		if err := acc.GetBuffer(&data, &availableFrameSize, &flags, &devicePosition, &qcpPosition); err != nil {
	//			continue
	//		}
	//		if availableFrameSize == 0 {
	//			continue
	//		}
	//
	//		start := unsafe.Pointer(data)
	//		lim := int(availableFrameSize) * int(wfx.NBlockAlign)
	//		buf := make([]byte, lim)
	//
	//		for n := 0; n < lim; n++ {
	//			b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
	//			buf[n] = *b
	//		}
	//
	//		go v.processBuf(buf)
	//
	//		offset += lim
	//
	//		if err := acc.ReleaseBuffer(availableFrameSize); err != nil {
	//			return err
	//		}
	//	}
	//}

	fmt.Println("stopping audio capture")
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

//func watchEvent(ctx context.Context, event uintptr) (err error) {
//	errorChan := make(chan error, 1)
//	go func() {
//		errorChan <- eventEmitter(event)
//	}()
//	select {
//	case err = <-errorChan:
//		close(errorChan)
//		return
//	case <-ctx.Done():
//		err = ctx.Err()
//		return
//	}
//	return
//}
//
//// this takes 10ms...
//func eventEmitter(event uintptr) (err error) {
//	//if err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
//	//	return
//	//}
//	dw := wca.WaitForSingleObject(event, wca.INFINITE)
//	if dw != 0 {
//		return fmt.Errorf("failed to watch event")
//	}
//	//ole.CoUninitialize()
//	return
//}

func (v *Visualizer) processBuf(buf []byte, peak float64, sampleRate float64) {
	//els := make([]float64, len(buf)/4)
	//sum := 0.0

	els := make([]float64, len(buf)/4)
	//sum := 0.0
	//max := 0.0
	//sum := 0.0
	for i := 0; i < len(buf); i += 4 {
		b := binary.LittleEndian.Uint32(buf[i : i+4])
		bi := int32(b)
		fbi := float64(bi)
		//fbi = math.Abs(fbi)
		//fbi /= 2
		els = append(els, fbi)
		//if math.Abs(fbi) > max {
		//	max = math.Abs(fbi)
		//}
		//sum += fbi
		//sum += math.Abs(fbi)
		//fbi := float64(bi)
		//els = append(els, float64(fbi))
		//sum += math.Abs(fbi)
	}

	//fmt.Println(els)
	//max = max / float64(len(els)) / math.MaxInt16 / 8
	//fmt.Print(max)

	//if max < peak {
	//	fmt.Println("max peak", max, peak)
	//}

	//avg := math.Pow(sum/float64(len(els))/math.MaxInt32*16, 2)
	//avg := sum / float64(len(els)) / math.MaxInt32 * 8
	//fmt.Print(avg)

	//fmt.Println(els)
	//avg := sum / float64(len(els))
	//fmt.Println(int(avg))

	//fmt.Println(avg / math.MaxInt32)

	avg := 0.0
	for _, el := range els {
		avg += math.Abs(el)
	}

	//peak = math.Min(peak, 1)
	avg = avg / float64(len(els)) / math.MaxInt32 * 8
	//avg = math.Min(avg, 1)

	//avg = math.Log2((avg+1)/2) + 1

	//avg *= 2
	//peak *= 2
	//peak *= 2

	avg = math.Min(avg, 1)
	//max = math.Min(max, 1)
	peak = math.Min(peak, 1)
	//max *= 10

	//max = math.Log2((max+1)/2) + 1
	//max = math.Log2((max+1)/2) + 1
	//max = math.Log2((max+1)/2) + 1
	//max = math.Log2((max+1)/2) + 1
	//max = math.Log2((max+1)/2) + 1

	//max = math.Min(max, 1)

	//peak = math.Log2((peak+1)/2) + 1

	//nrg =

	sig := dsp.Signal{
		SampleRate: sampleRate,
		Signal:     els,
	}

	//avg := 0.0
	normalized, err := sig.Normalize()
	if err != nil {
		//els = make([]float64, len(buf)/4)
	} else {
		//nrg := 0.0

		spectrum, err := normalized.FrequencySpectrum()
		if err != nil {
			fmt.Println(err)
			return
		}

		els = spectrum.Spectrum
	}

	pre := []float64{}

	//pre = append(pre, els[0])
	pre = append(pre, els[1])
	pre = append(pre, els[1])
	pre = append(pre, els[2])
	pre = append(pre, els[2])
	pre = append(pre, els[2])
	pre = append(pre, els[3])
	pre = append(pre, els[3])
	pre = append(pre, els[3])
	pre = append(pre, els[4])
	pre = append(pre, els[4])
	pre = append(pre, els[5])
	pre = append(pre, els[5])

	els = append(pre, els[6:]...)

	max := 0.0
	for _, el := range els {
		if math.Abs(el) > max {
			max = math.Abs(el)
		}
	}

	pix := []byte{}
	//pix = append(pix, []byte{0, 0, 0, 0xFF}...)
	//pix = append(pix, []byte{0, 0, 0, 0xFF}...)
	//pix = append(pix, []byte{0, 0, 0, 0xFF}...)
	//pix = append(pix, []byte{0, 0, 0, 0xFF}...)
	//pix = append(pix, []byte{0, 0, 0, 0xFF}...)
	//pix = append(pix, []byte{0, 0, 0, 0xFF}...)

	//fmt.Print("\n")

	//out := "\r"

	//avg *= 2
	//peak = math.Log2((peak+1)/2) + 1
	//peak = 1
	//avg = 1
	//avg = math.Log10(avg)

	next := 0.0
	for i := 0; i < v.maxLedCount; i++ {
		if i < len(els)-2 {
			next = els[i] + els[i+1]
		} else {
			next = els[i] + els[i]
		}
		curr := (next) / 2
		//curr := els[i]

		//curr *= 10
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1

		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//mult := float64(i / v.leds / 2)
		//offset := 250.0
		//loop := float64(360)
		//
		//hsv := colorful.Hsv(math.Mod((max+peak)*360+offset, loop), max, curr)

		norm := normalize(curr, 0, max)
		//h := curr * 360
		h := 0.0
		if norm < 0.2 {
			h = 320 // purple
		} else if norm < 0.7 {
			h = 240
		} else if norm < 0.8 {
			h = 320
		} else {
			h = 0
		}
		//h := 200. + norm*120

		h = math.Mod(h, 360)
		//fmt.Printf("%.32f\n", h)

		//s := math.Min((peak/3+0.66)/peak*2, 1)
		//s := 1.
		//s := math.Min(math.Abs((max+curr)*10), 1)
		s := (0.5 + (norm / 2))
		s = 1
		v := norm + 0.0
		v = (log((v+1)/2, 2) + 1)
		v *= (0.333 + peak*0.666)
		v = v*0.84 + 0.16

		//if max > .5 {
		//	s = 0
		//}

		//v := math.Min((curr + max), 1)

		hsv := colorful.Hsv(h, s, v)

		//curr = math.Log2((curr+1)/2) + 1
		//if avg > 0.9 {
		//	fmt.Println("avg", avg)
		//}
		//r, g, b := uint8(avg*curr*128+15), uint8(avg*curr*32+15), uint8(avg*curr*256+15)
		//rm := 1.0
		//gm := 1.0
		//bm := 1.0

		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1
		//curr = math.Log2((curr+1)/2) + 1

		//val := peak * curr * 256

		//curr *= 2

		//r, g, b := uint8(val), uint8(val*peak), uint8(val*peak)

		r, g, b, _ := hsv.RGBA()

		r = uint32(math.Min(float64(r/256), 255))
		g = uint32(math.Min(float64(g/256), 255))
		b = uint32(math.Min(float64(b/256), 255))

		//fmt.Println(r, g, b)
		//if r > 150 {
		//	r = 250
		//}

		//if r > max2 {
		//	max2 = r
		//}
		//
		//if g > max2 {
		//	max2 = g
		//}
		//
		//if b > max2 {
		//	max2 = b
		//}

		pix = append(pix, []byte{uint8(r), uint8(g), uint8(b), 0xFF}...)
		//pix = append(pix, []byte{uint8(r), uint8(g), uint8(b), 0xFF}...)

		//out += color.RGB(uint8(r), uint8(g), uint8(b), true).Sprintf(" ")
	}

	//fmt.Print(out)
	//fmt.Printf("%.4f  %.4f  %d  %.4f", avg, peak, max2, max)

	pixs = append(pixs, pix)
	if len(pixs) > 100 {
		pixs = pixs[1:]
	}

	weights := []float64{}
	weightsTotal := 0.0

	for i := 0; i < len(pixs); i++ {

		w := float64((i + 1) * (i + 1))
		//w := 1.0

		weights = append(weights, w)
		weightsTotal += w
	}
	//fmt.Println(weights)

	pix2 := make([]float64, len(pix))
	//fmt.Println("---", len(pixs))
	for i, p2 := range pixs {
		for j, p := range p2 {

			//fmt.Println(pix2[j])
			//fmt.Println(weights[i])

			pix2[j] = pix2[j] + float64(p)*weights[i]
			//pix2[j] = uint16(math.Max(float64(pix2[j]), float64(p)))
		}
	}

	pix3 := make([]float64, len(pix))
	for i, p := range pix2 {
		avg := p / weightsTotal
		//curr := p
		pix3[i] = float64(avg)
		//if i > 1 && i < len(pix2)-2 {
		//	avg = uint16(float64(float64(pix2[i+1]+avg) / float64(2.0)))
		//}
		//pix[i] = uint8(avg)
	}

	pix4 := make([]uint8, len(pix))

	for i := 0; i < len(pix3); i += 4 {
		offset := i

		if i >= len(pix3)/2 {
			offset = len(pix3) - 4 - i
		}

		pix4[i] = uint8(pix3[offset])
		pix4[i+1] = uint8(pix3[offset+1])
		pix4[i+2] = uint8(pix3[offset+2])
		pix4[i+3] = uint8(pix3[offset+3])

		//for i, p := range pix3 {
		//if i > 1 && i < len(pix3)-2 {
		//	pix4[i] = uint8(float64(float64(pix3[i+1]+p) / float64(2.0)))
		//} else {
		//	pix4[i] = uint8(p)
		//

		//pix4[i] = uint8(p)

	}

	//out := "\n"
	//for i := 0; i < len(pix); i += 4 {
	//	out += color.RGB(pix[i], pix[i+1], pix[i+2], true).Sprintf(" ")
	//}
	//fmt.Print(out)

	segs := []interfaces.Segment{}

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

			//for i, p := range pix3 {
			//if i > 1 && i < len(pix3)-2 {
			//	pix4[i] = uint8(float64(float64(pix3[i+1]+p) / float64(2.0)))
			//} else {
			//	pix4[i] = uint8(p)
			//

			//pix4[i] = uint8(p)

		}

		pix := pix4[:seg.Leds*4]

		if seg.Id == 0 {
			out := "\n"
			for i := 0; i < len(pix); i += 4 {
				out += color.RGB(pix[i], pix[i+1], pix[i+2], true).Sprintf(" ")
			}
			fmt.Print(out)
		}

		segs = append(segs, interfaces.Segment{
			Id:  seg.Id,
			Pix: pix,
		})

		//time.Sleep(100 * time.Nanosecond)
	}

	v.events <- interfaces.UpdateEvent{
		Segments: segs,
	}

	//pix = pix[6*4:]

	//v.events <- interfaces.UpdateEvent{
	//	SegmentId: 1,
	//	Data:      pix,
	//}
}

func log(v float64, base float64) float64 {
	return math.Log(v) / math.Log(base)
}

func normalize(val, min, max float64) float64 {
	return (val - min) / (max - min)
}

var pixs [][]byte

var max2 uint8

func New(opts ...Option) (*Visualizer, error) {
	v := &Visualizer{}

	for _, opt := range opts {
		err := opt(v)
		if err != nil {
			return nil, err
		}
	}

	v.events = make(chan interfaces.UpdateEvent, len(v.segments))

	return v, nil
}

type Segment struct {
	Id   int
	Leds int
}
