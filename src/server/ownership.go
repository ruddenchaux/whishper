package server

import (
	"context"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"

	"github.com/pluja/anysub/db"
	"github.com/pluja/anysub/ent/transcription"
	"github.com/pluja/anysub/ent/user"
	"github.com/pluja/anysub/utils/translations"
)

// isOwner reports whether transcription id belongs to the session user.
func isOwner(c iris.Context, id int) bool {
	uid, ok := sessions.Get(c).Get("user").(int)
	if !ok {
		return false
	}

	owned, err := db.Client().Transcription.Query().
		Where(transcription.IDEQ(id), transcription.HasUserWith(user.IDEQ(uid))).
		Exist(context.Background())
	return err == nil && owned
}

// RequireOwnership guards every route with an {id} transcription param:
// upstream handlers look transcriptions up by ID alone, so without this any
// logged-in user could read, edit or delete anyone's. Non-owners get a 404
// so other users' IDs don't leak. Must run after ApiMustBeLogged.
func (s *Server) RequireOwnership(c iris.Context) {
	if !c.Params().Exists("id") {
		c.Next()
		return
	}

	id, err := c.Params().GetInt("id")
	if err != nil || !isOwner(c, id) {
		c.StopWithStatus(iris.StatusNotFound)
		return
	}
	c.Next()
}

// RequireTranslations rejects translation routes when no LibreTranslate
// endpoint is configured, instead of failing on an empty URL.
func (s *Server) RequireTranslations(c iris.Context) {
	if !translations.Enabled() {
		c.StopWithText(iris.StatusServiceUnavailable, "translations are disabled: LIBRETRANSLATE_ENDPOINT is not set")
		return
	}
	c.Next()
}
