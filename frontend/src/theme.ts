import type { GlobalThemeOverrides } from 'naive-ui'

const FONT =
  'Inter, "Segoe UI", system-ui, -apple-system, "Microsoft YaHei", "PingFang SC", sans-serif'
const MONO = '"JetBrains Mono", ui-monospace, "SF Mono", "Cascadia Code", Consolas, monospace'

const shape: GlobalThemeOverrides = {
  common: {
    borderRadius: '8px',
    borderRadiusSmall: '6px',
    fontFamily: FONT,
    fontFamilyMono: MONO,
  },
  Button: { borderRadiusMedium: '8px', fontWeight: '600' },
  Input: { borderRadius: '8px' },
  InputNumber: { borderRadius: '8px' },
  Select: { borderRadius: '8px' },
  Radio: { buttonBorderRadius: '6px' },
  Modal: { borderRadius: '14px' },
}

export const lightThemeOverrides: GlobalThemeOverrides = {
  common: {
    ...shape.common,
    primaryColor: '#2563eb',
    primaryColorHover: '#3b82f6',
    primaryColorPressed: '#1d4ed8',
    primaryColorSuppl: '#3b82f6',
    infoColor: '#2563eb',
    successColor: '#16a34a',
    warningColor: '#f59e0b',
    errorColor: '#dc2626',
    textColorBase: '#172033',
    textColor1: '#172033',
    textColor2: '#3b4658',
    textColor3: '#6b7689',
    bodyColor: '#f6f7f9',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    popoverColor: '#ffffff',
    borderColor: '#d8dee8',
    dividerColor: '#edf1f6',
    inputColor: '#ffffff',
    inputColorDisabled: '#f2f4f7',
    placeholderColor: '#6b7689',
  },
  Button: shape.Button,
  Input: shape.Input,
  InputNumber: shape.InputNumber,
  Select: shape.Select,
  Radio: shape.Radio,
  Modal: shape.Modal,
}

export const darkThemeOverrides: GlobalThemeOverrides = {
  common: {
    ...shape.common,
    primaryColor: '#2b6cf0',
    primaryColorHover: '#4a87f5',
    primaryColorPressed: '#2458c4',
    primaryColorSuppl: '#4a87f5',
    infoColor: '#7aa8ff',
    successColor: '#34d399',
    warningColor: '#fbbf24',
    errorColor: '#f87171',
    textColorBase: '#e6eaf2',
    textColor1: '#e6eaf2',
    textColor2: '#b7c0d0',
    textColor3: '#8893a6',
    bodyColor: '#0f1218',
    cardColor: '#161b26',
    modalColor: '#161b26',
    popoverColor: '#1b2230',
    borderColor: '#2a3344',
    dividerColor: '#212a3a',
    inputColor: '#161b26',
    inputColorDisabled: '#1a2029',
    placeholderColor: '#8893a6',
  },
  Button: {
    ...shape.Button,
    textColorPrimary: '#e6eaf2',
    textColorHoverPrimary: '#ffffff',
    textColorPressedPrimary: '#ffffff',
    textColorFocusPrimary: '#ffffff',
    textColorDisabledPrimary: '#dddddd',
  },
  Input: shape.Input,
  InputNumber: shape.InputNumber,
  Select: shape.Select,
  Radio: shape.Radio,
  Modal: shape.Modal,
}
