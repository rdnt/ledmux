module ledctl3

go 1.18

require (
	github.com/eripe970/go-dsp-utils v0.0.0-20220123162022-4563116e558a
	github.com/go-ole/go-ole v1.2.6
	github.com/google/uuid v1.2.0
	github.com/gookit/color v1.5.0
	github.com/gorilla/websocket v1.4.2
	github.com/kbinani/screenshot v0.0.0-20210720154843-7d3a670d8329
	github.com/kirides/screencapture v0.0.0-20211031174040-89bc8578d816
	github.com/lithammer/shortuuid/v3 v3.0.7
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/moutend/go-wav v0.0.0-20170820031854-56127fbbb7ba
	github.com/moutend/go-wca v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/rdnt/go-scrap v0.0.0-20211023140534-1eef394fee02
	github.com/rpi-ws281x/rpi-ws281x-go v1.0.8
	github.com/sanity-io/litter v1.5.1
	github.com/vmihailenco/msgpack/v5 v5.3.5
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

require (
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/gen2brain/shm v0.0.0-20210511105953-083dbc7d9d83 // indirect
	github.com/goccmack/godsp v0.1.1 // indirect
	github.com/goccmack/goutil v0.4.0 // indirect
	github.com/jezek/xgb v0.0.0-20210312150743-0e0f116e1240 // indirect
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/mattetti/audio v0.0.0-20190404201502-c6aebeb78429 // indirect
	github.com/mjibson/go-dsp v0.0.0-20180508042940-11479a337f12 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
	golang.org/x/sys v0.0.0-20211031064116-611d5d643895 // indirect
)

replace github.com/kirides/screencapture v0.0.0-20211031174040-89bc8578d816 => ./pkg/screencapture
