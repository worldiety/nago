A Modal can be declared at any place in the composed view tree. However, these dialogs are teleported into
the modal space in tree declaration order. A Modal is layouted above all other regular content and if ModalTypeDialog
will catch focus and disable controls of the views behind. Its bounds are at most the maximum possible screen size.