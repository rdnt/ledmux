package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/eripe970/go-dsp-utils"
	"github.com/go-ole/go-ole"
	"github.com/gookit/color"
	"github.com/moutend/go-wav"
	"github.com/moutend/go-wca/pkg/wca"

	"ledctl3/cmd/audio/bpm"
)

var version = "latest"
var revision = "latest"

type DurationFlag struct {
	Value time.Duration
}

func (f *DurationFlag) Set(value string) (err error) {
	var sec float64

	if sec, err = strconv.ParseFloat(value, 64); err != nil {
		return
	}
	f.Value = time.Duration(sec * float64(time.Second))
	return
}

func (f *DurationFlag) String() string {
	return f.Value.String()
}

type FilenameFlag struct {
	Value string
}

func (f *FilenameFlag) Set(value string) (err error) {
	if !strings.HasSuffix(value, ".wav") {
		err = fmt.Errorf("specify WAVE audio file (*.wav)")
		return
	}
	f.Value = value
	return
}

func (f *FilenameFlag) String() string {
	return f.Value
}

func main() {
	var err error
	if err = run(os.Args); err != nil {
		panic(err)
	}
}

func run(args []string) (err error) {
	var durationFlag DurationFlag
	var filenameFlag FilenameFlag
	var versionFlag bool
	var audio *wav.File
	var file []byte

	f := flag.NewFlagSet(args[0], flag.ExitOnError)
	f.Var(&durationFlag, "duration", "Specify recording duration in second")
	f.Var(&durationFlag, "d", "Alias of --duration")
	f.Var(&filenameFlag, "output", "file name")
	f.Var(&filenameFlag, "o", "Alias of --output")
	f.BoolVar(&versionFlag, "version", false, "Show version")
	f.Parse(args[1:])

	if versionFlag {
		fmt.Printf("%s-%s\n", version, revision)
		return
	}
	//if filenameFlag.Value == "" {
	//
	//	return
	//}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case <-signalChan:
			fmt.Println("Interrupted by SIGINT")
			cancel()
		}
		return
	}()

	if audio, err = captureSharedEventDriven(ctx, durationFlag.Value); err != nil {
		return
	}
	if file, err = wav.Marshal(audio); err != nil {
		return
	}
	if err = ioutil.WriteFile(filenameFlag.Value, file, 0644); err != nil {
		return
	}
	fmt.Println("Successfully done")
	return
}

func captureSharedEventDriven(ctx context.Context, duration time.Duration) (audio *wav.File, err error) {
	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	if err = mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return
	}
	defer mmd.Release()

	var ps *wca.IPropertyStore
	if err = mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		return
	}
	defer ps.Release()

	var pv wca.PROPVARIANT
	if err = ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		return
	}
	fmt.Printf("Capturing audio from: %s\n", pv.String())

	var ac *wca.IAudioClient
	if err = mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &ac); err != nil {
		return
	}
	defer ac.Release()

	var wfx *wca.WAVEFORMATEX
	if err = ac.GetMixFormat(&wfx); err != nil {
		return
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	wfx.WFormatTag = 1
	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
	wfx.CbSize = 0

	if audio, err = wav.New(int(wfx.NSamplesPerSec), int(wfx.WBitsPerSample), int(wfx.NChannels)); err != nil {
		return
	}

	fmt.Println("--------")
	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	fmt.Printf("Channels: %d\n", wfx.NChannels)
	fmt.Println("--------")

	var defaultPeriod wca.REFERENCE_TIME
	var minimumPeriod wca.REFERENCE_TIME
	var latency time.Duration
	if err = ac.GetDevicePeriod(&defaultPeriod, &minimumPeriod); err != nil {
		return
	}
	latency = time.Duration(int(minimumPeriod) * 100)

	fmt.Println("Default period: ", defaultPeriod)
	fmt.Println("Minimum period: ", minimumPeriod)
	fmt.Println("Latency: ", latency)

	if err = ac.Initialize(wca.AUDCLNT_SHAREMODE_SHARED, wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK|wca.AUDCLNT_STREAMFLAGS_LOOPBACK, defaultPeriod, 0, wfx, nil); err != nil {
		return
	}

	audioReadyEvent := wca.CreateEventExA(0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE)
	defer wca.CloseHandle(audioReadyEvent)

	if err = ac.SetEventHandle(audioReadyEvent); err != nil {
		return
	}

	var bufferFrameSize uint32
	if err = ac.GetBufferSize(&bufferFrameSize); err != nil {
		return
	}
	fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)

	var acc *wca.IAudioCaptureClient
	if err = ac.GetService(wca.IID_IAudioCaptureClient, &acc); err != nil {
		return
	}
	defer acc.Release()

	if err = ac.Start(); err != nil {
		return
	}
	fmt.Println("Start capturing with shared event driven mode")
	if duration <= 0 {
		fmt.Println("Press Ctrl-C to stop capturing")
	}

	var output = []byte{}
	var offset int
	var isCapturing = true
	var currentDuration time.Duration
	var b *byte
	var data *byte
	var availableFrameSize uint32
	var flags uint32
	var devicePosition uint64
	var qcpPosition uint64
	//var padding uint32

	errorChan := make(chan error, 1)

	//in := make(chan float32)
	//out := make(chan float32)

	//go bpm.ProgressivelyReadFloatArray(in, out)

	//done := make(chan bool)

	//go readProgressiveVars(out, done, *progressive, *progressiveInterval)

	//var scores []float64

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
		case err = <-errorChan:
			currentDuration = time.Duration(float64(offset) / float64(wfx.WBitsPerSample/8) / float64(wfx.NChannels) / float64(wfx.NSamplesPerSec) * float64(time.Second))
			if duration != 0 && currentDuration > duration {
				isCapturing = false
				break
			}
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

			//fmt.Println(data, availableFrameSize, flags, devicePosition, qcpPosition)

			start := unsafe.Pointer(data)
			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
			buf := make([]byte, lim)

			//fmt.Println(len(buf), cap(buf))

			for n := 0; n < lim; n++ {
				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
				buf[n] = *b
			}

			// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

			go processBuf(buf)

			offset += lim
			//output = append(output, buf...)

			if err = acc.ReleaseBuffer(availableFrameSize); err != nil {
				return
			}
		}
	}

	io.Copy(audio, bytes.NewBuffer(output))
	fmt.Println("Stop capturing")

	if err = ac.Stop(); err != nil {
		return
	}
	return
}

func processBuf(buf []byte) {
	els := make([]float64, len(buf)/4)
	sum := 0.0

	for i := 0; i < len(buf); i += 4 {
		b := binary.LittleEndian.Uint32(buf[i : i+4])
		bi := int32(b)
		fbi := float64(bi)
		els = append(els, float64(fbi))
		sum += math.Abs(fbi)
	}

	//fmt.Println(els)
	//avg := sum / float64(len(els))
	//fmt.Println(int(avg))

	//fmt.Println(avg / math.MaxInt32)

	sig := dsp.Signal{
		SampleRate: 96000,
		Signal:     els,
	}

	normalized, _ := sig.Normalize()

	spectrum, _ := normalized.FrequencySpectrum()

	els = spectrum.Spectrum

	//fmt.Print("\r")
	fmt.Print("\n")

	// loop only the first 8th	q
	for i := 0; i < len(els)/16; i++ {
		c := uint8((els[i]) * 256)
		color.RGB(c, c, c, true).Print(" ")
	}
	//fmt.Print("\n")
	//for i := 0; i < len(els)/16; i++ {
	//	c := uint8((els[i]) * 256)
	//	color.RGB(c, c, c, true).Print(" ")
	//}
}

// ZScore on 16bit WAV samples
func ZScore(samples []float64, lag int, threshold float64, influence float64) (signals []float64) {
	//lag := 20
	//threshold := 3.5
	//influence := 0.5

	signals = make([]float64, len(samples))
	filteredY := make([]float64, len(samples))
	for i, sample := range samples[0:lag] {
		filteredY[i] = sample
	}
	avgFilter := make([]float64, len(samples))
	stdFilter := make([]float64, len(samples))

	avgFilter[lag] = Average(samples[0:lag])
	stdFilter[lag] = Std(samples[0:lag])

	for i := lag + 1; i < len(samples); i++ {

		f := samples[i]

		if math.Abs(samples[i]-avgFilter[i-1]) > threshold*stdFilter[i-1] {
			if samples[i] > avgFilter[i-1] {
				signals[i] = 1
			} else {
				signals[i] = -1
			}
			filteredY[i] = influence*f + (1-influence)*filteredY[i-1]
			avgFilter[i] = Average(filteredY[(i - lag):i])
			stdFilter[i] = Std(filteredY[(i - lag):i])
		} else {
			signals[i] = 0
			filteredY[i] = samples[i]
			avgFilter[i] = Average(filteredY[(i - lag):i])
			stdFilter[i] = Std(filteredY[(i - lag):i])
		}
	}

	return
}

func Std(num []float64) float64 {
	var sum, mean, sd float64
	for i := 1; i <= len(num); i++ {
		sum += num[i-1]
	}
	mean = sum / float64(len(num))

	for j := 0; j < len(num); j++ {
		// The use of Pow math function func Pow(x, y float64) float64
		sd += math.Pow(num[j]-mean, 2)
	}
	// The use of Sqrt math function func Sqrt(x float64) float64
	sd = math.Sqrt(sd / 10)

	return sd
}

func Average(chunk []float64) (avg float64) {
	var sum float64
	for _, sample := range chunk {
		if sample < 0 {
			sample *= -1
		}
		sum += sample
	}
	return sum / float64(len(chunk))
}

var (
	min                 = flag.Float64("min", 84, "min BPM you are expecting")
	max                 = flag.Float64("max", 146, "max BPM you are expecting")
	progressive         = flag.Bool("progressive", true, "Print the BPM for every period")
	progressiveInterval = flag.Int("interval", 10, "How many seconds for every progressive chunk printed")
)

func readProgressiveVars(input chan float32, done chan bool, progressive bool, pint int) {
	if progressive {
		maxsize := calcChunkLen(pint)

		nrg := make([]float32, 0)
		for f := range input {
			nrg = append(nrg, f)
			if len(nrg) == maxsize {
				fmt.Printf("%f\n", bpm.ScanForBpm(nrg, *min, *max, 1024, 20000))
				nrg = make([]float32, 0)
			}
		}
		done <- true
	} else {
		nrg := make([]float32, 0)
		for f := range input {
			nrg = append(nrg, f)
		}
		fmt.Printf("%f\n", bpm.ScanForBpm(nrg, *min, *max, 1024, 20000))
		done <- true
	}
}

func calcChunkLen(second int) int {
	return (bpm.RATE / bpm.INTERVAL) * second
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
