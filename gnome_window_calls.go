/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) 2018 Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	// Interface của extension window-calls (https://github.com/ickyicky/window-calls).
	// Vài extension khác cũng expose đúng interface này nên chỉ probe theo
	// interface, đừng probe theo uuid của extension.
	windowCallsPath   = dbus.ObjectPath("/org/gnome/Shell/Extensions/Windows")
	windowCallsMethod = "org.gnome.Shell.Extensions.Windows.List"

	// FocusIn gọi hàm này đồng bộ nên phải có hạn, không để gnome-shell treo
	// thì treo luôn cả việc gõ.
	windowCallsTimeout = time.Second
)

// window-calls đặt tên field là `focus` (nó build từ `props.has = ['focus']`),
// còn một số extension tương thích lại đặt là `has_focus` — nhận cả hai.
type windowCallsWindow struct {
	WmClass         string `json:"wm_class"`
	WmClassInstance string `json:"wm_class_instance"`
	Focus           *bool  `json:"focus"`
	HasFocus        *bool  `json:"has_focus"`
}

func (w *windowCallsWindow) hasFocus() bool {
	if w.Focus != nil {
		return *w.Focus
	}
	return w.HasFocus != nil && *w.HasFocus
}

// wmClass ghép lại thành "instance:class" cho giống WM_CLASS mà đường X11 trả
// về, để InputModeMapping trong config dùng chung được giữa X11 và Wayland.
func (w *windowCallsWindow) wmClass() string {
	if w.WmClassInstance == "" {
		return w.WmClass
	}
	if w.WmClass == "" {
		return w.WmClassInstance
	}
	return w.WmClassInstance + ":" + w.WmClass
}

func parseWindowCallsFocusClass(payload string) (string, error) {
	var windows []windowCallsWindow
	if err := json.Unmarshal([]byte(payload), &windows); err != nil {
		return "", err
	}
	for i := range windows {
		if windows[i].hasFocus() {
			if wmClass := windows[i].wmClass(); wmClass != "" {
				return wmClass, nil
			}
			return "", errors.New("focused window has no wm class")
		}
	}
	return "", fmt.Errorf("no focused window among %d window(s)", len(windows))
}

// windowCallsGetFocusWindowClass lấy wm class của cửa sổ đang focus qua extension
// window-calls. Khác đường org.gnome.Shell.Eval, nó không cần bật unsafe mode; và
// vì mutter set wm_class của cửa sổ Wayland từ app_id nên app native Wayland cũng
// có wm class, tức chọn input mode theo từng app hoạt động lại được (xem #540).
func windowCallsGetFocusWindowClass() (string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return "", err
	}
	// Không Close: SessionBus trả về connection dùng chung của cả process.
	var ctx, cancel = context.WithTimeout(context.Background(), windowCallsTimeout)
	defer cancel()

	var payload string
	var obj = conn.Object("org.gnome.Shell", windowCallsPath)
	if err := obj.CallWithContext(ctx, windowCallsMethod, 0).Store(&payload); err != nil {
		return "", err
	}
	return parseWindowCallsFocusClass(payload)
}
