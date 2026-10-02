---
title: Feedback & Overlays
weight: 4
sidebar:
  open: false
---

Feedback components tell the user what happened, e.g. that a save succeeded or an error occurred. Overlay
components show content in a layer above the view, like dialogs and popovers. Most of them live in package
`presentation/ui/alert`.

**Messages**

- [Banner](banner/): a message inside the view, with an intent like success or error
- [Banner Messages](banner_messages/): short-lived notifications on top of the view
- [Banner Error](banner_error/): an error as a banner, without leaking details

**Dialogs and overlays**

- [Dialog](dialog/): a modal dialog with standard buttons
- [Dialog Create](dialog_create/) and [Dialog Edit](dialog_edit/): dialogs with a generated form
- [Custom Dialog](custom_dialog/): the plain dialog box for your own layouts
- [Modal](modal/): the layer below dialogs, popovers and notifications

**Errors**

- [Error View](error_view/): replaces a view which failed to render
- [Support Request Dialog](support_request_dialog/): reports unexpected technical errors

**Navigation**

- [Breadcrumbs](breadcrumbs/): the path to the current page
