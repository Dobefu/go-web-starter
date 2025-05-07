const beforeUnloadHandler = (e: Event): void => {
  e.preventDefault()
}

const formElements = document.querySelectorAll<HTMLInputElement>(
  'input,textarea,[contenteditable]',
)

formElements.forEach((formElement) => {
  const initialFormValue = formElement.value

  formElement.addEventListener('input', (e: Event) => {
    if (
      e.target &&
      'value' in e.target &&
      e.target.value !== initialFormValue
    ) {
      addEventListener('beforeunload', beforeUnloadHandler)
      return
    }

    removeEventListener('beforeunload', beforeUnloadHandler)
  })
})
