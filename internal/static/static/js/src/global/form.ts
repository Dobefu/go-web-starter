const beforeUnloadHandler = (e: Event): void => {
  e.preventDefault()
}

const formElements = document.querySelectorAll<HTMLInputElement>(
  ':is(input:not([type=search],[type=hidden]),textarea,[contenteditable]):not(data-ignore-dirty)',
)

formElements.forEach((formElement) => {
  const isCheckable =
    formElement.type === 'checkbox' || formElement.type === 'radio'
  const initialFormValue = isCheckable ? formElement.checked : formElement.value

  formElement.addEventListener('input', (e: Event) => {
    if (!(e.target instanceof HTMLInputElement)) {
      return
    }

    const val = isCheckable ? e.target.checked : e.target.value

    if (val !== initialFormValue) {
      addEventListener('beforeunload', beforeUnloadHandler)
      return
    }

    removeEventListener('beforeunload', beforeUnloadHandler)
  })
})
