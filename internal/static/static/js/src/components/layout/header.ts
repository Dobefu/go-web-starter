;(function initializeMobileMenu(): void {
  const mobileMenuToggle = document.getElementById('mobile-menu--toggle')
  const mobileMenuClose = document.getElementById('mobile-menu--close')
  const mobileMenuBackdrop = document.getElementById('mobile-menu--backdrop')
  const mobileMenu = document.getElementById('mobile-menu')

  const toggleMenu = (): void => {
    if (!mobileMenu) {
      return
    }

    mobileMenu.setAttribute(
      'aria-expanded',
      mobileMenu.getAttribute('aria-expanded') === 'true' ? 'false' : 'true',
    )
  }

  mobileMenuToggle?.addEventListener('click', toggleMenu)
  mobileMenuClose?.addEventListener('click', toggleMenu)
  mobileMenuBackdrop?.addEventListener('click', toggleMenu)
})()
