package event

type SetLedsEvent struct {
	Event
	Segments []SetLedsEventSegment `json:"segments"`
}

type SetLedsEventSegment struct {
	Id  int    `json:"id"`
	Pix []byte `json:"pix"`
}

func NewSetLedsEvent(segments []SetLedsEventSegment) SetLedsEvent {
	return SetLedsEvent{
		Event: Event{
			Type: Render,
		},
		Segments: segments,
	}
}
