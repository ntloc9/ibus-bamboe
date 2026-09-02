package main

import "testing"

func TestParseWindowCallsFocusClass(t *testing.T) {
	for _, tc := range []struct {
		name     string
		payload  string
		expected string
		hasError bool
	}{
		{
			// window-calls build field focus từ props.has = ['focus'].
			name: "window_calls_focus_field",
			payload: `[{"wm_class":"Navigator","wm_class_instance":"Navigator","title":"t","focus":false},
				{"wm_class":"Gnome-terminal","wm_class_instance":"gnome-terminal-server","title":"t","focus":true}]`,
			expected: "gnome-terminal-server:Gnome-terminal",
		},
		{
			// Một số extension tương thích đặt tên field là has_focus.
			name: "compatible_extension_has_focus_field",
			payload: `[{"wm_class":"Gnome-terminal","wm_class_instance":"gnome-terminal-server","has_focus":false},
				{"wm_class":"Code","wm_class_instance":"code","has_focus":true}]`,
			expected: "code:Code",
		},
		{
			// App native Wayland: mutter lấy cả hai field từ app_id.
			name:     "wayland_app_id",
			payload:  `[{"wm_class":"microsoft-edge","wm_class_instance":"microsoft-edge","focus":true}]`,
			expected: "microsoft-edge:microsoft-edge",
		},
		{
			name:     "instance_missing",
			payload:  `[{"wm_class":"Code","wm_class_instance":"","focus":true}]`,
			expected: "Code",
		},
		{
			name:     "no_focused_window",
			payload:  `[{"wm_class":"Code","wm_class_instance":"code","focus":false}]`,
			hasError: true,
		},
		{
			name:     "focused_window_without_wm_class",
			payload:  `[{"wm_class":"","wm_class_instance":"","focus":true}]`,
			hasError: true,
		},
		{
			name:     "empty_list",
			payload:  `[]`,
			hasError: true,
		},
		{
			name:     "invalid_json",
			payload:  `not json`,
			hasError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wmClass, err := parseWindowCallsFocusClass(tc.payload)
			if tc.hasError {
				if err == nil {
					t.Errorf("Expected an error, got wm class (%s).", wmClass)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %s.", err)
			}
			if wmClass != tc.expected {
				t.Errorf("Wm class, expected (%s), got (%s).", tc.expected, wmClass)
			}
		})
	}
}
