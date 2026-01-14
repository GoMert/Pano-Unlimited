package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var currentVariant = theme.VariantDark

// Light colors
var (
	lightBg      = color.RGBA{R: 243, G: 243, B: 243, A: 255}
	lightSurface = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	lightBorder  = color.RGBA{R: 220, G: 220, B: 220, A: 255}
	lightText    = color.RGBA{R: 32, G: 32, B: 32, A: 255}
	lightTextSec = color.RGBA{R: 96, G: 96, B: 96, A: 255}
	lightPrimary = color.RGBA{R: 0, G: 120, B: 212, A: 255} // Windows blue
	lightPinned  = color.RGBA{R: 255, G: 249, B: 230, A: 255}
	lightPinBrd  = color.RGBA{R: 255, G: 200, B: 100, A: 255}
)

// Dark colors
var (
	darkBg      = color.RGBA{R: 32, G: 32, B: 32, A: 255}
	darkSurface = color.RGBA{R: 43, G: 43, B: 43, A: 255}
	darkBorder  = color.RGBA{R: 60, G: 60, B: 60, A: 255}
	darkText    = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	darkTextSec = color.RGBA{R: 180, G: 180, B: 180, A: 255}
	darkPrimary = color.RGBA{R: 96, G: 205, B: 255, A: 255}
	darkPinned  = color.RGBA{R: 50, G: 45, B: 35, A: 255}
	darkPinBrd  = color.RGBA{R: 180, G: 140, B: 60, A: 255}
)

type PanoTheme struct {
	variant fyne.ThemeVariant
}

func NewLightTheme() fyne.Theme {
	currentVariant = theme.VariantLight
	return &PanoTheme{variant: theme.VariantLight}
}

func NewDarkTheme() fyne.Theme {
	currentVariant = theme.VariantDark
	return &PanoTheme{variant: theme.VariantDark}
}

func IsDarkMode() bool {
	return currentVariant == theme.VariantDark
}

func GetCardBackgroundColor(pinned bool) color.Color {
	if IsDarkMode() {
		if pinned {
			return darkPinned
		}
		return darkSurface
	}
	if pinned {
		return lightPinned
	}
	return lightSurface
}

func GetCardBorderColor(pinned bool) color.Color {
	if IsDarkMode() {
		if pinned {
			return darkPinBrd
		}
		return darkBorder
	}
	if pinned {
		return lightPinBrd
	}
	return lightBorder
}

func GetTextColor() color.Color {
	if IsDarkMode() {
		return darkText
	}
	return lightText
}

func GetSecondaryTextColor() color.Color {
	if IsDarkMode() {
		return darkTextSec
	}
	return lightTextSec
}

func GetPrimaryColor() color.Color {
	if IsDarkMode() {
		return darkPrimary
	}
	return lightPrimary
}

func GetBadgeColors(badgeType string) (bg color.Color, fg color.Color) {
	if IsDarkMode() {
		return darkSurface, darkTextSec
	}
	return lightSurface, lightTextSec
}

func (t *PanoTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	v := t.variant

	switch name {
	case theme.ColorNameBackground:
		if v == theme.VariantDark {
			return darkBg
		}
		return lightBg

	case theme.ColorNameButton:
		if v == theme.VariantDark {
			return color.RGBA{R: 55, G: 55, B: 55, A: 255}
		}
		return color.RGBA{R: 240, G: 240, B: 240, A: 255}

	case theme.ColorNameForeground:
		if v == theme.VariantDark {
			return darkText
		}
		return lightText

	case theme.ColorNamePrimary:
		if v == theme.VariantDark {
			return darkPrimary
		}
		return lightPrimary

	case theme.ColorNameHover:
		if v == theme.VariantDark {
			return color.RGBA{R: 65, G: 65, B: 65, A: 255}
		}
		return color.RGBA{R: 230, G: 230, B: 230, A: 255}

	case theme.ColorNameSeparator:
		if v == theme.VariantDark {
			return darkBorder
		}
		return lightBorder

	case theme.ColorNameInputBackground:
		if v == theme.VariantDark {
			return darkSurface
		}
		return lightSurface

	case theme.ColorNameScrollBar:
		if v == theme.VariantDark {
			return color.RGBA{R: 80, G: 80, B: 80, A: 255}
		}
		return color.RGBA{R: 200, G: 200, B: 200, A: 255}

	case theme.ColorNameDisabled:
		if v == theme.VariantDark {
			return color.RGBA{R: 100, G: 100, B: 100, A: 255}
		}
		return color.RGBA{R: 160, G: 160, B: 160, A: 255}

	default:
		return theme.DefaultTheme().Color(name, v)
	}
}

func (t *PanoTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *PanoTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *PanoTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 16
	case theme.SizeNameScrollBar:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}
