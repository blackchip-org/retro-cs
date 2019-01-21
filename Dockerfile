FROM ubuntu:bionic

RUN apt-get update
RUN apt-get install -y \
    golang \
    git \
    libsdl2-dev \
    libsdl2-image-dev \
    libsdl2-mixer-dev \
    libsdl2-ttf-dev \
    libsdl2-gfx-dev

RUN mkdir -p /root/go/src/github.com/blackchip-org/retro-cs
RUN go get github.com/veandco/go-sdl2/sdl
RUN go get github.com/chzyer/readline
