import { Workbox } from 'workbox-window'

let wb: any

if ('serviceWorker' in navigator && process.env.NODE_ENV === 'production') {
  wb = new Workbox(`${process.env.BASE_URL}service-worker.js`)

  wb.addEventListener('controlling', () => {
    window.location.reload()
  })

  wb.register()
} else {
  wb = null
}

export default wb
