package server

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"

	"github.com/pluja/anysub/db"
	"github.com/pluja/anysub/ent"
	"github.com/pluja/anysub/ent/transcription"
	"github.com/pluja/anysub/ent/user"
	"github.com/pluja/anysub/frontend"
	"github.com/pluja/anysub/utils"
)

// ownTranscription loads transcription {id} if it belongs to the session
// user; anyone else gets a 404 so IDs of other users' files don't leak.
func ownTranscription(c iris.Context) (*ent.Transcription, bool) {
	id, err := c.Params().GetInt("id")
	if err != nil {
		c.StopWithError(iris.StatusBadRequest, errors.New("invalid id"))
		return nil, false
	}

	uid, ok := sessions.Get(c).Get("user").(int)
	if !ok {
		c.Redirect("/login")
		return nil, false
	}

	tx, err := db.Client().Transcription.Query().
		Where(transcription.IDEQ(id), transcription.HasUserWith(user.IDEQ(uid))).
		WithTranslations().
		Only(context.Background())
	if err != nil {
		c.StopWithStatus(iris.StatusNotFound)
		return nil, false
	}
	return tx, true
}

func (s *Server) TranscriptionPageHandler(c iris.Context) {
	tx, ok := ownTranscription(c)
	if !ok {
		return
	}

	if err := c.RenderComponent(frontend.TranscriptionPage(tx)); err != nil {
		c.HTML("<h3>%s</h3>", err.Error())
	}
}

func (s *Server) TranscriptionAudioHandler(c iris.Context) {
	tx, ok := ownTranscription(c)
	if !ok {
		return
	}

	// filepath.Base guards against a stored name containing path segments.
	src := filepath.Join(utils.Getenv("UPLOAD_DIR", "/app/uploads"), filepath.Base(tx.FileName))
	// ServeFile goes through http.ServeContent, so Range requests (seeking) work.
	if err := c.ServeFile(src); err != nil {
		c.StopWithStatus(iris.StatusNotFound)
	}
}
