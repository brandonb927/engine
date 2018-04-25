package filestore

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

type mockWriter struct {
	text   string
	err    error
	closed bool
}

func (w *mockWriter) WriteString(s string) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	w.text += s
	return len(s), nil
}

func (w *mockWriter) Close() error {
	w.closed = true
	return nil
}

var basicGame = &pb.Game{
	ID:           "myid",
	Status:       "asdf",
	Width:        10,
	Height:       15,
	SnakeTimeout: 200,
	TurnTimeout:  100,
	Mode:         "multiplayer",
}

var deadSnake = &pb.Snake{
	ID:   "snake1",
	Name: "snake1",
	URL:  "http://snake1",
	Body: []*pb.Point{
		&pb.Point{X: 4, Y: 4}, &pb.Point{X: 4, Y: 3},
	},
	Death: &pb.Death{
		Cause: "death-cause",
		Turn:  1,
	},
	Color: "red",
}

var basicSnakes = []*pb.Snake{
	&pb.Snake{
		ID:   "snake1",
		Name: "snake1",
		URL:  "http://snake1",
		Body: []*pb.Point{
			&pb.Point{X: 4, Y: 4}, &pb.Point{X: 4, Y: 3},
		},
		Death: nil,
		Color: "red",
	},
	&pb.Snake{
		ID:   "snake2",
		Name: "snake2",
		URL:  "http://snake2",
		Body: []*pb.Point{
			&pb.Point{X: 6, Y: 4}, &pb.Point{X: 6, Y: 3},
		},
		Death: nil,
		Color: "green",
	},
}

var basicTicks = []*pb.GameTick{
	&pb.GameTick{
		Turn:   1,
		Food:   []*pb.Point{&pb.Point{X: 1, Y: 1}},
		Snakes: basicSnakes,
	},
	&pb.GameTick{
		Turn:   2,
		Food:   []*pb.Point{&pb.Point{X: 1, Y: 1}},
		Snakes: basicSnakes,
	},
}

var tickWithDeadSnake = &pb.GameTick{
	Turn:   1,
	Food:   []*pb.Point{&pb.Point{X: 1, Y: 1}},
	Snakes: []*pb.Snake{deadSnake},
}

func checkBasicGameJSON(t *testing.T, j string) {
	info := gameInfo{}
	err := json.Unmarshal([]byte(j), &info)
	require.NoError(t, err)

	require.Equal(t, "myid", info.ID)
	require.Equal(t, int64(10), info.Width)
	require.Equal(t, int64(15), info.Height)
	require.Len(t, info.Snakes, 2)
	require.Equal(t, "snake1", info.Snakes[0].ID)
	require.Equal(t, "snake2", info.Snakes[1].ID)
}

func checkBasicFrameJSON(t *testing.T, j string, turn int64) {
	f := frame{}
	err := json.Unmarshal([]byte(j), &f)
	require.NoError(t, err)

	require.Equal(t, turn, f.Turn, "wrong turn")
	require.Len(t, f.Food, 1, "wrong food count")
	require.Equal(t, int64(1), f.Food[0].X)
	require.Equal(t, int64(1), f.Food[0].Y)
	require.Len(t, f.Snakes, 2)
	require.Equal(t, "snake1", f.Snakes[0].ID)
	require.Equal(t, "snake2", f.Snakes[1].ID)
	require.Nil(t, f.Snakes[0].Death)
	require.Nil(t, f.Snakes[1].Death)
}

func checkDeadSnakeFrameJSON(t *testing.T, j string) {
	f := frame{}
	err := json.Unmarshal([]byte(j), &f)
	require.NoError(t, err)

	require.Len(t, f.Snakes, 1)
	require.Equal(t, "snake1", f.Snakes[0].ID)
	require.NotNil(t, f.Snakes[0].Death)
	require.Equal(t, "death-cause", f.Snakes[0].Death.Cause)
	require.Equal(t, int64(1), f.Snakes[0].Death.Turn)
}

func TestWriteGameInfo(t *testing.T) {
	w := &mockWriter{
		closed: false,
	}
	err := writeGameInfo(w, basicGame, basicSnakes)
	require.NoError(t, err)
	checkBasicGameJSON(t, w.text)
}

func TestWriteGameInfoError(t *testing.T) {
	w := &mockWriter{
		err:    errors.New("fail"),
		closed: false,
	}
	err := writeGameInfo(w, basicGame, basicSnakes)
	require.NotNil(t, err)
}

func TestWriteTick(t *testing.T) {
	w := &mockWriter{
		closed: false,
	}
	err := writeTick(w, basicTicks[0])
	require.NoError(t, err)
	checkBasicFrameJSON(t, w.text, 1)
}

func TestWriteTickDeadSnake(t *testing.T) {
	w := &mockWriter{
		closed: false,
	}
	err := writeTick(w, tickWithDeadSnake)
	require.NoError(t, err)
	checkDeadSnakeFrameJSON(t, w.text)
}

func TestWriteTickError(t *testing.T) {
	w := &mockWriter{
		err:    errors.New("fail"),
		closed: false,
	}
	err := writeTick(w, basicTicks[0])
	require.NotNil(t, err)
}
