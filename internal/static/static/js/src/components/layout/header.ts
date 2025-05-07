const mobileMenuToggle = document.getElementById('mobile-menu--toggle')
const mobileMenuClose = document.getElementById('mobile-menu--close')
const mobileMenuBackdrop = document.getElementById('mobile-menu--backdrop')
const mobileMenu = document.getElementById('mobile-menu')

const menuControls = [mobileMenu, mobileMenuClose, mobileMenuBackdrop]
const skipIds = ['mobile-menu--close', 'mobile-menu--backdrop']

const setInert = (elements: (Element | null)[], value: boolean): void => {
  elements.forEach((el) => {
    if (!el) return
    value ? el.setAttribute('inert', '') : el.removeAttribute('inert')
  })
}

const updateMenuInertState = (): void => {
  const isOpen = mobileMenu?.getAttribute('aria-expanded') === 'true'

  if (window.innerWidth <= 768) {
    setInert(menuControls, !isOpen)
    setInert([mobileMenuToggle], isOpen)
  } else {
    setInert([...menuControls, mobileMenuToggle], false)
  }
}

addEventListener('resize', updateMenuInertState)
addEventListener('DOMContentLoaded', updateMenuInertState)

const isOrContainsMenu = (el: Element, menu: Element): boolean =>
  el === menu || el.contains(menu)

const toggleMenu = (): void => {
  if (!mobileMenu) {
    return
  }

  const isOpen = mobileMenu.getAttribute('aria-expanded') === 'true'
  const allPageChildren = [
    ...Array.from(document.body.children),
    ...Array.from(document.querySelectorAll('header > nav > *')),
  ]

  if (isOpen) {
    mobileMenu.setAttribute('aria-expanded', 'false')
    mobileMenu.removeAttribute('inert')
    allPageChildren.forEach((el) => {
      if (
        !isOrContainsMenu(el, mobileMenu) &&
        !skipIds.includes(el.id) &&
        el.tagName !== 'SCRIPT'
      ) {
        el.removeAttribute('inert')
      }
    })

    updateMenuInertState()
    return
  }

  mobileMenu.setAttribute('aria-expanded', 'true')
  mobileMenu.removeAttribute('inert')
  allPageChildren.forEach((el) => {
    if (
      !isOrContainsMenu(el, mobileMenu) &&
      !skipIds.includes(el.id) &&
      el.tagName !== 'SCRIPT'
    ) {
      el.setAttribute('inert', '')
    }
  })

  updateMenuInertState()
}

;[mobileMenuToggle, mobileMenuClose, mobileMenuBackdrop].forEach((btn) =>
  btn?.addEventListener('click', toggleMenu),
)
