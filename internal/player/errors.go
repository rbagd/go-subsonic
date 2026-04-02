package player

import "errors"

var ErrAudioDisabled = errors.New("audio disabled (built with -tags noaudio)")
