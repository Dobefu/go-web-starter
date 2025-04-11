// Mobile menu functionality
export const init = (): void => {
  const menuButton = document.querySelector(
    '[aria-controls="mobile-menu"]',
  ) as HTMLButtonElement
  const mobileMenu = document.getElementById('mobile-menu')
  const backdrop = document.getElementById('mobile-menu-backdrop')
  const menuPanel = document.getElementById('mobile-menu-panel')
  const closeButton = menuPanel?.querySelector(
    'button[aria-label="Close menu"]',
  ) as HTMLButtonElement
  const menuLinks = menuPanel?.querySelectorAll('a') || []
  const focusableElements = [closeButton, ...menuLinks].filter(
    Boolean,
  ) as HTMLElement[]
  const firstFocusableElement = focusableElements[0]
  const lastFocusableElement = focusableElements[focusableElements.length - 1]

  if (!menuButton || !mobileMenu || !backdrop || !menuPanel || !closeButton) {
    return
  }

  const trapFocus = (e: KeyboardEvent): void => {
    if (e.key !== 'Tab') return

    if (e.shiftKey) {
      if (document.activeElement === firstFocusableElement) {
        e.preventDefault()
        lastFocusableElement.focus()
      }
    } else {
      if (document.activeElement === lastFocusableElement) {
        e.preventDefault()
        firstFocusableElement.focus()
      }
    }
  }

  const toggleMenu = (): void => {
    const isExpanded = menuButton.getAttribute('aria-expanded') === 'true'
    menuButton.setAttribute('aria-expanded', (!isExpanded).toString())

    if (!isExpanded) {
      mobileMenu.classList.remove('hidden')
      // Trigger reflow
      mobileMenu.offsetHeight
      backdrop.classList.remove('opacity-0')
      menuPanel.classList.remove('-translate-x-full')
      // Focus the first element
      firstFocusableElement.focus()
      // Add focus trap
      menuPanel.addEventListener('keydown', trapFocus)
    } else {
      backdrop.classList.add('opacity-0')
      menuPanel.classList.add('-translate-x-full')
      // Remove focus trap
      menuPanel.removeEventListener('keydown', trapFocus)
      // Return focus to menu button
      menuButton.focus()
      // Wait for transition to complete before hiding
      setTimeout(() => {
        if (backdrop.classList.contains('opacity-0')) {
          mobileMenu.classList.add('hidden')
        }
      }, 300)
    }

    document.body.style.overflow = !isExpanded ? 'hidden' : ''
  }

  menuButton.addEventListener('click', toggleMenu)
  backdrop.addEventListener('click', toggleMenu)
  closeButton.addEventListener('click', toggleMenu)

  // Close menu when clicking a link
  menuLinks.forEach((link) => {
    link.addEventListener('click', toggleMenu)
  })
}
