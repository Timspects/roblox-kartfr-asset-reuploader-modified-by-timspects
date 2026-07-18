package context

import (
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/app/response"
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/roblox"
)

type Context struct {
	Client          *roblox.Client
	Logger          *logger
	PauseController *pauseController
	Response        *response.Response
}

func New(c *roblox.Client, resp *response.Response) *Context {
	return &Context{
		Client:          c,
		Logger:          newLogger(),
		PauseController: newPauseController(),
		Response:        resp,
	}
}
