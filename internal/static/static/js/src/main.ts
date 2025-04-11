type Module = { init?: () => void }
const modules = import.meta.glob<Module>('./components/**/*.ts')

Object.values(modules).forEach(async (module) => {
  const mod = await module()

  if (mod && typeof mod.init === 'function') {
    mod.init()
  }
})
