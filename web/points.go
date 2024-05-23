package web

import (
	"encoding/binary"
	"fmt"
	"net/http"

	"forzatelemetry/models"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) points(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	streamer := newProtobufStreamer(w, r)
	for point, err := range h.db.IterPoints(id, nil, r.Context()) {
		if err != nil {
			streamer.Fail(err)
			return
		}
		streamer.Send(point.ToProto())
	}
}

type protobufStreamer struct {
	w        http.ResponseWriter
	r        *http.Request
	dataSent bool
}

func newProtobufStreamer(w http.ResponseWriter, r *http.Request) *protobufStreamer {
	return &protobufStreamer{
		w:        w,
		r:        r,
		dataSent: false,
	}
}

func (rd *protobufStreamer) Send(v *models.ApiPoint) error {
	data, err := proto.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed marshaling to protobuf: %v", err)
	}

	if !rd.dataSent {
		rd.dataSent = true
		rd.w.Header().Set("Content-Type", "application/protobuf")
		rd.w.WriteHeader(http.StatusOK)
	}

	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	rd.w.Write(size)
	rd.w.Write(data)
	return nil
}

func (rd *protobufStreamer) Fail(err error) {
	if !rd.dataSent {
		rd.dataSent = true
		Render(rd.w, rd.r, StorageErrorRenderer(err))
	}
}
