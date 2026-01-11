package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Theme variants
var (
	currentVariant = theme.VariantLight
)

// Light theme colors
var (
	// Light - Primary colors
	lightPrimaryColor = color.RGBA{R: 59, G: 130, B: 246, A: 255}  // Modern blue
	lightPrimaryHover = color.RGBA{R: 37, G: 99, B: 235, A: 255}   // Darker blue

	// Light - Background colors
	lightBgColor      = color.RGBA{R: 250, G: 250, B: 252, A: 255} // Off-white
	lightSurfaceColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // Pure white

	// Light - Text colors
	lightTextPrimary   = color.RGBA{R: 17, G: 24, B: 39, A: 255}    // Near black
	lightTextSecondary = color.RGBA{R: 107, G: 114, B: 128, A: 255} // Gray

	// Light - Border color
	lightBorderColor = color.RGBA{R: 229, G: 231, B: 235, A: 255} // Light gray
)

// Dark theme colors
var (
	// Dark - Primary colors
	darkPrimaryColor = color.RGBA{R: 96, G: 165, B: 250, A: 255}  // Lighter blue for dark mode
	darkPrimaryHover = color.RGBA{R: 59, G: 130, B: 246, A: 255}  // Standard blue

	// Dark - Background colors
	darkBgColor      = color.RGBA{R: 17, G: 24, B: 39, A: 255}    // Very dark blue-gray
	darkSurfaceColor = color.RGBA{R: 30, G: 41, B: 59, A: 255}    // Slate-800

	// Dark - Text colors
	darkTextPrimary   = color.RGBA{R: 248, G: 250, B: 252, A: 255} // Near white
	darkTextSecondary = color.RGBA{R: 148, G: 163, B: 184, A: 255} // Light gray

	// Dark - Border color
	darkBorderColor = color.RGBA{R: 51, G: 65, B: 85, A: 255} // Slate-700
)

// Common accent colors
var (
	successColor = color.RGBA{R: 34, G: 197, B: 94, A: 255}  // Green
	dangerColor  = color.RGBA{R: 239, G: 68, B: 68, A: 255}  // Red
	warningColor = color.RGBA{R: 245, G: 158, B: 11, A: 255} // Amber
)

// PanoTheme represents the application theme
type PanoTheme struct {
	variant fyne.ThemeVariant
}

// NewLightTheme creates a modern light theme
func NewLightTheme() fyne.Theme {
	currentVariant = theme.VariantLight
	return &PanoTheme{variant: theme.VariantLight}
}

// NewDarkTheme creates a modern dark theme
func NewDarkTheme() fyne.Theme {
	currentVariant = theme.VariantDark
	return &PanoTheme{variant: theme.VariantDark}
}

// IsDarkMode returns true if dark mode is active
func IsDarkMode() bool {
	return currentVariant == theme.VariantDark
}

// GetCardBackgroundColor returns the appropriate card background color
func GetCardBackgroundColor(pinned bool) color.RGBA {
	if IsDarkMode() {
		if pinned {
			return color.RGBA{R: 55, G: 48, B: 23, A: 255} // Dark amber/yellow tint
		}
		return darkSurfaceColor
	}
	if pinned {
		return color.RGBA{R: 254, G: 252, B: 232, A: 255} // Subtle yellow
	}
	return lightSurfaceColor
}

// GetCardBorderColor returns the appropriate card border color
func GetCardBorderColor() color.RGBA {
	if IsDarkMode() {
		return darkBorderColor
	}
	return lightBorderColor
}

// GetTextColor returns the appropriate text color
func GetTextColor() color.RGBA {
	if IsDarkMode() {
		return darkTextPrimary
	}
	return lightTextPrimary
}

// GetSecondaryTextColor returns the appropriate secondary text color
func GetSecondaryTextColor() color.RGBA {
	if IsDarkMode() {
		return darkTextSecondary
	}
	return lightTextSecondary
}

// Color returns theme colors
func (t *PanoTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Use internal variant
	v := t.variant

	switch name {
	case theme.ColorNameBackground:
		if v == theme.VariantDark {
			return darkBgColor
		}
		return lightBgColor

	case theme.ColorNameButton:
		if v == theme.VariantDark {
			return darkSurfaceColor
		}
		return lightSurfaceColor

	case theme.ColorNameForeground:
		if v == theme.VariantDark {
			return darkTextPrimary
		}
		return lightTextPrimary

	case theme.ColorNamePrimary:
		if v == theme.VariantDark {
			return darkPrimaryColor
		}
		return lightPrimaryColor

	case theme.ColorNameHover:
		if v == theme.VariantDark {
			return color.RGBA{R: 51, G: 65, B: 85, A: 255} // Slate-700
		}
		return color.RGBA{R: 243, G: 244, B: 246, A: 255}

	case theme.ColorNameFocus:
		if v == theme.VariantDark {
			return color.RGBA{R: 96, G: 165, B: 250, A: 80}
		}
		return color.RGBA{R: 59, G: 130, B: 246, A: 80}

	case theme.ColorNameDisabled:
		if v == theme.VariantDark {
			return color.RGBA{R: 100, G: 116, B: 139, A: 255} // Slate-500
		}
		return lightTextSecondary

	case theme.ColorNamePlaceHolder:
		if v == theme.VariantDark {
			return darkTextSecondary
		}
		return lightTextSecondary

	case theme.ColorNameSeparator:
		if v == theme.VariantDark {
			return darkBorderColor
		}
		return lightBorderColor

	case theme.ColorNameInputBackground:
		if v == theme.VariantDark {
			return color.RGBA{R: 30, G: 41, B: 59, A: 255} // Slate-800
		}
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}

	case theme.ColorNameScrollBar:
		if v == theme.VariantDark {
			return color.RGBA{R: 71, G: 85, B: 105, A: 255} // Slate-600
		}
		return color.RGBA{R: 203, G: 213, B: 225, A: 255} // Slate-300

	case theme.ColorNameSuccess:
		return successColor
	case theme.ColorNameError:
		return dangerColor
	case theme.ColorNameWarning:
		return warningColor

	default:
		return theme.DefaultTheme().Color(name, v)
	}
}

// Font returns theme fonts
func (t *PanoTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns theme icons
func (t *PanoTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns theme sizes
func (t *PanoTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInlineIcon:
		return 18
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 16
	default:
		return theme.DefaultTheme().Size(name)
	}
}
