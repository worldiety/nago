---
title: Basic
weight: 1
sidebar:
  open: false
---

Basic components are the small building blocks of a view: they display content or take a single input. Most of
them live in package `presentation/ui`; a few have their own package below it.

**Actions**

- [Button](button/): primary, secondary and tertiary buttons
- [Filled Button](filled_button/): a button with a custom background color

**Content**

- [Text](text/): headlines, paragraphs and links
- [Rich Text](rich_text/): formatted HTML content
- [Image](image/): images and icons
- [QR Code](qr_code/): a scannable code for a string
- [Colored Text Pill](colored_text_pill/): tags and status badges
- [Card](card/): a titled block of content
- [WebView](webview/), [Video](video/) and [PDF](pdf/): embedded media

**Inputs**

- [Text Field](text_field/): text and number input
- [Checkbox](checkbox/) and [Checkbox Field](checkbox_field/)
- [Radio Button](radio_button/) and [Radio Button Field](radiobutton_field/)
- [Toggle](toggle/): a switch with immediate effect
- [Select](select/): a single choice from a list
- [Slider](slider/): a number or a range

Inputs bind to a `*core.State`, see the setters `InputValue` or `InputChecked`. Also pass the current value of the
state to the constructor, e.g. `Checkbox(checked.Get()).InputValue(checked)`: most inputs display the constructor
value, not the state.
