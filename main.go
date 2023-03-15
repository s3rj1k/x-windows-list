package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/jezek/xgb/xproto"
)

const (
	FontSize   = 18.0
	FontFamily = "Monospace"

	WindowWidth  = 960
	WindowHeight = 540
)

const (
	ColumnWID = iota
	ColumnWName
)

func main() {
	gtk.Init(nil)
	log.SetFlags(0)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	treeView, err := gtk.TreeViewNew()
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	cellRndrTxtWID, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	cellRndrTxtWName, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	for _, el := range []*gtk.CellRendererText{cellRndrTxtWID, cellRndrTxtWName} {
		err = el.SetProperty("family", FontFamily)
		if err != nil {
			log.Fatalf("GTK: %v\n", err)
		}

		err = el.SetProperty("size-points", FontSize)
		if err != nil {
			log.Fatalf("GTK: %v\n", err)
		}

		err = el.SetProperty("size-set", true)
		if err != nil {
			log.Fatalf("GTK: %v\n", err)
		}
	}

	err = cellRndrTxtWName.SetProperty("ellipsize", pango.ELLIPSIZE_END)
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	colWID, err := gtk.TreeViewColumnNewWithAttribute("WID", cellRndrTxtWID, "text", ColumnWID)
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	colWName, err := gtk.TreeViewColumnNewWithAttribute("WNAME", cellRndrTxtWName, "text", ColumnWName)
	if err != nil {
		log.Fatalf("GTK: %v\n", err)
	}

	selection, err := treeView.GetSelection()
	if err != nil {
		log.Fatalf("GTK: %v", err)
	}

	win.SetDecorated(false)
	win.SetDefaultSize(WindowWidth, WindowHeight)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetResizable(false)
	win.SetSkipPagerHint(true)
	win.SetSkipTaskbarHint(true)
	win.SetTitle("X Windows List")
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_NORMAL)
	win.Add(scroll)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(ev)
		if keyEvent.KeyVal() == gdk.KEY_Escape {
			win.Close()
		}
	})

	scroll.Add(treeView)

	treeView.AppendColumn(colWID)
	treeView.AppendColumn(colWName)
	treeView.SetHeadersVisible(false)
	treeView.SetModel(listStore)
	treeView.SetActivateOnSingleClick(true)
	treeView.SetGridLines(gtk.TREE_VIEW_GRID_LINES_HORIZONTAL)
	treeView.SetSearchColumn(ColumnWName)
	treeView.SetHoverSelection(true)

	colWID.SetVisible(false)

	selection.SetMode(gtk.SELECTION_SINGLE)

	xwl, err := List()
	if err != nil {
		log.Fatalf("X: %v\n", err)
	}

	for _, el := range xwl {
		if el.IsListable() {
			iter := listStore.Append()

			err := listStore.Set(
				iter,
				[]int{
					ColumnWID, ColumnWName,
				},
				[]interface{}{
					el.GetHumanReadableID(),
					el.Name,
				},
			)
			if err != nil {
				log.Fatalf("GTK: %v\n", err)
			}
		}
	}

	treeView.ConnectAfter("row-activated", func(tv *gtk.TreeView, path *gtk.TreePath) {
		ts, err := tv.GetSelection()
		if err != nil {
			log.Fatalf("GTK: %s\n", err)
		}

		itm, iter, ok := ts.GetSelected()
		if !ok {
			return
		}

		tm := itm.ToTreeModel()

		gval, err := tm.GetValue(iter, ColumnWID)
		if err != nil {
			log.Fatalf("GTK: %s\n", err)
		}

		val, err := gval.GetString()
		if err != nil {
			log.Fatalf("GTK: %s\n", err)
		}

		ui64, err := strconv.ParseUint(
			strings.ReplaceAll(val, "0x", ""),
			16, 32,
		)
		if err != nil {
			log.Fatalf("GO: %s\n", err)
		}

		err = SetActiveWindow(xproto.Window(ui64))
		if err != nil {
			log.Fatalf("X: %s\n", err)
		}

		win.Close()
	})

	win.ShowAll()

	gtk.Main()
}
