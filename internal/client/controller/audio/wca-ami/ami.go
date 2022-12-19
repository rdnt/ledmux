// Package wca_ami includes interfaces for the AudioMeterInformation API.
// Taken from https://github.com/moutend/go-wca/pull/11
package wca_ami

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
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
