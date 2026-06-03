function initChoices(context) {
  var scope = context || document
  scope.querySelectorAll('select.item-product-select').forEach(function(el) {
    if (typeof Choices === 'undefined') return
    try {
      if (el._choices) {
        try { el._choices.destroy() } catch(e) {}
        delete el._choices
      }
    } catch(e) {}
    try {
      el._choices = new Choices(el, {
        searchEnabled: true,
        searchPlaceholderValue: 'Digite para filtrar...',
        shouldSort: false,
        itemSelectText: '',
      })
    } catch(e) {
      console.warn('Choices init failed:', e)
    }
  })
}

document.addEventListener('DOMContentLoaded', function() { initChoices() })
document.addEventListener('htmx:afterSwap', function(e) {
  var target = e.detail && e.detail.target
  if (target && target.querySelector) {
    if (target.closest('#patomanco') || target.id === 'patomanco' || target.querySelector('#patomanco')) {
      initChoices(target)
    }
  }
})
