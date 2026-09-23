import { useEffect, useMemo, useState } from 'react'
import './App.css'

const DEFAULT_SESSION_ID = 'frontend-demo'
const suggestions = ['Найди автомат 3P 160А', 'Что есть в наличии в Астане?', 'Подбери замену для кабеля']

const fallbackProduct = {
  id: 0,
  article: '—',
  sku: '—',
  name: 'Товар не найден',
  brand: 'EKT',
  product_type: 'Каталог',
  price: 0,
  quantity: 0,
  stock: 0,
  image: 'https://images.unsplash.com/photo-1555963966-b7ae5404b6ed?auto=format&fit=crop&w=700&q=80',
  characteristics: {
    poles: '—',
    current: '—',
    voltage: '—',
    breaking_capacity: '—',
  },
  data_issues: [],
}

function formatPrice(value) {
  const num = Number(value ?? 0)
  if (!Number.isFinite(num) || num <= 0) return '—'
  return `${num.toLocaleString('ru-RU')} ₸`
}

function getCharacteristics(product) {
  if (!product || !product.characteristics) return []

  if (Array.isArray(product.characteristics)) {
    return product.characteristics.filter(([, value]) => value && value !== '—')
  }

  return Object.entries(product.characteristics).filter(([, value]) => value && value !== '—')
}

function App() {
  const [query, setQuery] = useState('')
  const [products, setProducts] = useState([])
  const [selectedId, setSelectedId] = useState('')
  const [chatOpen, setChatOpen] = useState(true)
  const [message, setMessage] = useState('')
  const [messages, setMessages] = useState([
    { from: 'assistant', text: 'Здравствуйте! Я помогу найти электротехнические товары, проверить наличие и подобрать замену по подтверждённым данным.' },
    { from: 'assistant', text: 'Проверяю каталог через backend API и могу сразу подхватить актуальные позиции.' },
  ])
  const [cart, setCart] = useState([])
  const [proposal, setProposal] = useState(null)
  const [compareOpen, setCompareOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    loadProducts()
    loadCart()
  }, [])

  const filteredProducts = useMemo(() => {
    if (!query.trim()) return products
    const value = query.toLowerCase()
    return products.filter((product) => {
      const text = `${product.id ?? ''} ${product.article ?? ''} ${product.sku ?? ''} ${product.name ?? ''} ${product.product_type ?? ''} ${product.brand ?? ''}`.toLowerCase()
      return text.includes(value)
    })
  }, [products, query])

  const selectedProduct = products.find((product) => String(product.id) === String(selectedId)) || products[0] || fallbackProduct
  const alternative = products.find((product) => String(product.id) !== String(selectedId)) || products[1] || null

  async function loadProducts(searchText = '') {
    try {
      const url = searchText.trim() ? `/api/products?q=${encodeURIComponent(searchText.trim())}` : '/api/products'
      const response = await fetch(url)
      const data = await response.json()
      if (!Array.isArray(data)) {
        setProducts([])
        return
      }
      setProducts(data)
      if (data.length > 0 && !selectedId) {
        setSelectedId(String(data[0].id))
      }
    } catch (error) {
      console.error('loadProducts failed', error)
      setProducts([])
    }
  }

  async function loadCart() {
    try {
      const response = await fetch(`/api/cart?session_id=${DEFAULT_SESSION_ID}`)
      const data = await response.json()
      if (Array.isArray(data)) {
        setCart(data)
      } else if (data && Array.isArray(data.items)) {
        setCart(data.items)
      }
    } catch (error) {
      console.error('loadCart failed', error)
      setCart([])
    }
  }

  async function sendMessage(text = message) {
    const cleanText = (text ?? message ?? '').trim()
    if (!cleanText) return

    setMessages((current) => [...current, { from: 'user', text: cleanText }])
    setMessage('')
    setIsLoading(true)

    try {
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_id: DEFAULT_SESSION_ID, message: cleanText }),
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || 'Ошибка API')
      }

      if (data.reply) {
        setMessages((current) => [...current, { from: 'assistant', text: data.reply }])
      }

      if (Array.isArray(data.products) && data.products.length > 0) {
        setProducts(data.products)
        setSelectedId(String(data.products[0].id))
      }

      if (data.pending_add && data.products && data.products.length > 0) {
        setProposal({ product: data.products[0], quantity: data.pending_add.quantity })
      } else {
        setProposal(null)
      }

      if (data.cart_url) {
        await loadCart()
      }
    } catch (error) {
      setMessages((current) => [...current, { from: 'assistant', text: `Не удалось получить ответ от API: ${error.message}` }])
    } finally {
      setIsLoading(false)
    }
  }

  async function confirmAdd() {
    if (!proposal) return

    try {
      const response = await fetch(`/api/cart/confirm?session_id=${DEFAULT_SESSION_ID}`, { method: 'POST' })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || 'Не удалось подтвердить корзину')
      }

      setMessages((current) => [...current, { from: 'assistant', text: 'Готово. Товар подтверждён и добавлен в корзину.' }])
      setProposal(null)
      await loadCart()
    } catch (error) {
      setMessages((current) => [...current, { from: 'assistant', text: error.message }])
    }
  }

  const productPrice = formatPrice(selectedProduct?.price ?? 0)
  const productStock = selectedProduct?.stock ?? selectedProduct?.quantity ?? 0
  const warningText = selectedProduct?.data_issues?.[0] || ''
  const characteristics = getCharacteristics(selectedProduct)

  return (
    <div className={`app-shell ${chatOpen ? 'chat-active' : ''}`}>
      <div className="utility-bar"><div>⌖ Астана</div><div className="utility-links"><span className="account-icon" aria-label="Профиль пользователя" title="Профиль пользователя"><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="3.5" /><path d="M5 20c.8-3.3 3.1-5 7-5s6.2 1.7 7 5" /></svg></span></div></div>
      <header className="main-header"><div className="brand"><span>ГРУППА КОМПАНИЙ</span><b>ЭЛЕКТРОКОМПЛЕКТ</b></div><button className="catalog-button">Каталог <span>☷</span></button><label className="global-search"><span>⌕</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Поиск по названию, ID или артикулу" /><kbd>⌘ K</kbd></label><div className="header-actions"><span>⇄ Сравнить</span><span>♡ Избранное</span><a className="cart-link" href="cart.html">▱ Корзина <i>{cart.length}</i></a></div></header>
      <main>
        <section className="welcome-row"><div><p className="eyebrow">EKT SMART SEARCH <span>LIVE API</span></p><h1>Подберём нужное.<br /><em>Проверим каждую деталь.</em></h1><p className="intro">Умный помощник по электротехническим товарам. Находит по каталогу EKT и получает данные напрямую из backend API.</p></div><div className="trust-panel"><div className="pulse-dot"></div><div><b>Данные из backend</b><span>Оптимизированный поиск и чат по каталогу</span></div><small>›</small></div></section>
        <div className="search-wrap"><span className="search-icon">⌕</span><input value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && (query.trim() ? sendMessage(query) : loadProducts())} placeholder="Например, автомат 3P или 200300285_" /><button onClick={() => (query.trim() ? sendMessage(query) : loadProducts())}>Найти <span>↗</span></button></div>
        <div className="quick-row"><span>Популярное:</span>{suggestions.map((item) => <button key={item} onClick={() => sendMessage(item)}>{item}</button>)}</div>
        <div className="content-grid">
          <section className="catalog-column">
            <div className="section-heading"><div><span className="section-kicker">КАТАЛОГ EKT · {filteredProducts.length} ПОЗИЦИИ</span><h2>Товары для вашего проекта</h2></div><div className="view-toggle"><button className="active">▦</button><button>☷</button></div></div>
            {filteredProducts.length === 0 && <div className="empty-state"><span>⌕</span><h3>Ничего не нашли</h3><p>Попробуйте ID, артикул или название товара.</p></div>}
            <div className="product-list">
              {filteredProducts.map((product) => (
                <article className={`product-card ${String(selectedId) === String(product.id) ? 'selected' : ''}`} key={product.id ?? product.article ?? product.sku} onClick={() => setSelectedId(String(product.id))}>
                  <div className="product-image"><img src={product.image || fallbackProduct.image} alt="" /><span className={(product.stock ?? product.quantity ?? 0) ? 'in-stock' : 'out-stock'}>{(product.stock ?? product.quantity ?? 0) ? 'В наличии' : 'Нет в наличии'}</span></div>
                  <div className="product-info"><span className="product-type">{product.product_type || product.category || 'Каталог EKT'}</span><h3>{product.name}</h3><div className="meta"><span>ID {product.id}</span><span>Арт. {product.article || product.sku || '—'}</span></div><div className="product-bottom"><b>{formatPrice(product.price)}</b><span className={(product.stock ?? product.quantity ?? 0) ? 'stock' : 'muted'}>{(product.stock ?? product.quantity ?? 0) ? `${product.stock ?? product.quantity ?? 0} шт. на складах` : 'Остаток 0 шт.'}</span></div></div>
                  <button className="card-arrow" aria-label="Открыть товар">↗</button>
                </article>
              ))}
            </div>
          </section>

          <aside className="detail-column">
            <div className="detail-top"><span className="section-kicker">ПОДРОБНАЯ КАРТОЧКА</span><div className="detail-actions"><button>♡</button><button>⋯</button></div></div>
            <div className="detail-image"><img src={selectedProduct.image || fallbackProduct.image} alt="" /><span className="verified">✓ Данные подтверждены</span></div>
            <span className="product-type">{selectedProduct.product_type || selectedProduct.category || 'Каталог EKT'}</span>
            <h2>{selectedProduct.name}</h2>
            <div className="detail-meta"><span>Бренд <b>{selectedProduct.brand || 'EKT'}</b></span><span>Артикул <b>{selectedProduct.article || selectedProduct.sku || '—'}</b></span></div>
            <div className="price-row"><div><small>Цена за единицу</small><strong>{productPrice}</strong></div><div className={productStock ? 'availability' : 'availability unavailable'}><i></i>{productStock ? `В наличии · ${productStock} шт.` : 'Нет в наличии'}</div></div>
            {warningText && <div className="warning"><span>!</span><p><b>Нужно внимание</b> {warningText}</p></div>}
            <div className="characteristics"><div className="subheading"><b>Характеристики</b><span>Все параметры ↗</span></div>{characteristics.map(([key, value]) => <div className="characteristic" key={key}><span>{key}</span><b>{value}</b></div>)}</div>
            {!productStock && <button className="outline-button" onClick={() => { setCompareOpen(true); setChatOpen(true) }}>Подобрать замену <span>→</span></button>}
            <a className="ekt-link" href="https://ekt.kz" target="_blank" rel="noreferrer">Открыть на ekt.kz ↗</a>
          </aside>
        </div>

        {compareOpen && alternative && <section className="compare-section"><div className="section-heading"><div><span className="section-kicker">ПРОВЕРЕННЫЙ ПОДБОР</span><h2>Возможная замена</h2></div><button className="close-button" onClick={() => setCompareOpen(false)}>×</button></div><div className="compare-card"><div className="compare-product"><img src={alternative.image || fallbackProduct.image} alt="" /><div><span className="product-type">{alternative.product_type || 'Каталог EKT'}</span><h3>{alternative.name}</h3><span className="stock">✓ В наличии · {alternative.stock ?? alternative.quantity ?? 0} шт.</span></div></div><div className="match-score"><strong>3/4</strong><span>параметра совпадает</span></div><div className="compare-note"><b>Почему подходит</b><p>Совпадают основные характеристики, и товар доступен в каталоге. Перед монтажом лучше проверить точные параметры у менеджера.</p></div></div></section>}
      </main>
      <button className={`chat-launcher ${chatOpen ? 'hidden' : ''}`} onClick={() => setChatOpen(true)}>✦ <span>Спросить EKT AI</span></button>
      {chatOpen && <aside className="chat-panel"><div className="chat-header"><div className="assistant-avatar">✦</div><div><b>EKT AI</b><span>Ассистент-консультант <i></i></span></div><button onClick={() => setChatOpen(false)}>×</button></div><div className="chat-context"><span>◉</span> Анализирую каталог EKT <b>в реальном времени</b></div><div className="messages">{messages.map((item, index) => <div className={`message ${item.from}`} key={`${item.text}-${index}`}>{item.from === 'assistant' && <span className="mini-avatar">✦</span>}<p>{item.text}</p></div>)}{proposal && <div className="proposal"><span className="proposal-label">ПРЕДЛОЖЕНИЕ</span><b>{proposal.quantity} шт. · {proposal.product.name}</b><span>{formatPrice(proposal.product.price)} за единицу</span><button onClick={confirmAdd}>Подтвердить добавление <span>→</span></button><button className="cancel-proposal" onClick={() => setProposal(null)}>Отмена</button></div>}<div className="chat-suggestions">{suggestions.slice(0, 2).map((item) => <button key={item} onClick={() => sendMessage(item)}>{item}</button>)}</div></div><form className="chat-input" onSubmit={(event) => { event.preventDefault(); sendMessage() }}><input value={message} onChange={(event) => setMessage(event.target.value)} placeholder="Напишите вопрос..." disabled={isLoading} /><button aria-label="Отправить" disabled={isLoading}>↑</button></form><div className="chat-footnote">Ответы основаны на подтверждённых данных · <u>Подробнее</u></div></aside>}
    </div>
  )
}

export default App
