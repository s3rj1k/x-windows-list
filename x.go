package main

import (
	"fmt"
	"strings"

	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
	"github.com/jezek/xgbutil/icccm"
	"github.com/jezek/xgbutil/xgraphics"
)

const (
	MaxIconWidth  = 32
	MaxIconHeight = 32
)

type XWindowList struct {
	Icon  *xgraphics.Image
	Name  string
	Type  []string
	State []string
	ID    xproto.Window
}

func (xw *XWindowList) GetHumanReadableID() string {
	return fmt.Sprintf("0x%x", xw.ID)
}

func (xw *XWindowList) IsListable() bool {
	for _, el := range xw.Type {
		if el == "_NET_WM_WINDOW_TYPE_DOCK" {
			return false
		}
	}

	for _, el := range xw.State {
		if el == "_NET_WM_STATE_SKIP_TASKBAR" {
			return false
		}

		if el == "_NET_WM_STATE_SKIP_PAGER" {
			return false
		}
	}

	return true
}

func List() ([]XWindowList, error) {
	x, err := xgbutil.NewConn()
	if err != nil {
		return nil, err //nolint:wrapcheck // pass error unwrapped
	}

	defer x.Conn().Close()

	ids, err := ewmh.ClientListGet(x)
	if err != nil {
		return nil, err //nolint:wrapcheck // pass error unwrapped
	}

	out := make([]XWindowList, 0, len(ids))

	for _, id := range ids {
		var xw XWindowList

		xw.ID = id
		xw.Name = GetXWName(x, id)
		xw.Type, _ = ewmh.WmWindowTypeGet(x, id)
		xw.State, _ = ewmh.WmStateGet(x, id)
		xw.Icon, _ = xgraphics.FindIcon(x, id, MaxIconWidth, MaxIconHeight)

		out = append(out, xw)
	}

	return out, nil
}

func GetXWName(x *xgbutil.XUtil, id xproto.Window) string {
	name, err := ewmh.WmNameGet(x, id)
	if err != nil || strings.TrimSpace(name) == "" {
		name, err = icccm.WmNameGet(x, id)
		if err != nil || strings.TrimSpace(name) == "" {
			name = "N/A"
		}
	}

	return name
}

func SetActiveWindow(id xproto.Window) error {
	x, err := xgbutil.NewConn()
	if err != nil {
		return err //nolint:wrapcheck // pass error unwrapped
	}

	defer x.Conn().Close()

	return ewmh.ActiveWindowReq(x, id) //nolint:wrapcheck // pass error unwrapped
}
