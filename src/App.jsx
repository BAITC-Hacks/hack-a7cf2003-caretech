import { useMemo, useState } from 'react'
import { api } from './api'
import './App.css'

const demoProducts = [
  { id: '515291', article: '200300285_', name: 'Автоматический выключатель ВА47-100 3P 160А', brand: 'EKF', type: 'Низковольтная аппаратура', price: '64920', quantity: 23, image: 'https://images.unsplash.com/photo-1555963966-b7ae5404b6ed?auto=format&fit=crop&w=700&q=80', data_issues: ['Номинальный ток: 160 А в названии и 250 А в свойстве'], characteristics: [['Полюса', '3P'], ['Напряжение', '400 В'], ['Ток', 'Требует уточнения'], ['Отключающая способность', '10 кА']] },
  { id: '48783', article: '200100442_', name: 'Выключатель автоматический ВА47-100 3P 100А', brand: 'EKF', type: 'Низковольтная аппаратура', price: '42700', quantity: 12, image: 'https://images.unsplash.com/photo-1621905252507-b35492cc74b4?auto=format&fit=crop&w=700&q=80', characteristics: [['Полюса', '3P'], ['Напряжение', '400 В'], ['Ток', '100 А'], ['Отключающая способность', '10 кА']] },
  { id: '23466', article: '300410120_', name: 'Кабель силовой ВВГнг(А)-LS 3х2.5', brand: 'ЭКСПЕРТ-КАБЕЛЬ', type: 'Кабель / Провод', price: '385', quantity: 0, image: 'https://images.unsplash.com/photo-1617897711388-0480c3d7d6d8?auto=format&fit=crop&w=700&q=80', characteristics: [['Жилы', '3'], ['Сечение', '2.5 мм²'], ['Напряжение', '0.66 кВ'], ['Длина', 'Бухта']] },
]
const suggestions = ['Есть ли 515291?', 'Чем заменить, если нет?', 'Как купить?']

function normalizeProduct(item) {
  if (!item) return null
  const characteristics = item.characteristics && !Array.isArray(item.characteristics) ? Object.entries(item.characteristics) : item.characteristics || []
  return { ...item, id: String(item.id), quantity: item.quantity ?? null, characteristics }
}

function priceLabel(price) {
  if (price === null || price === undefined || price === '') return 'Нет данных'
  return String(price)
}

function App() {
  const [demoMode, setDemoMode] = useState(import.meta.env.VITE_DEMO_MODE === 'true')
  const [query, setQuery] = useState('')
  const [products, setProducts] = useState(demoMode ? demoProducts : [])
  const [selectedId, setSelectedId] = useState(demoMode ? '515291' : null)
  const [selectedProduct, setSelectedProduct] = useState(demoMode ? demoProducts[0] : null)
  const [messages, setMessages] = useState(demoMode ? [{ from: 'assistant', text: 'Демо-режим включён. Задайте вопрос, например «Есть ли 515291?».' }] : [{ from: 'assistant', text: 'Здравствуйте! Начните с вопроса «Есть ли 515291?» — я обращусь к проверенным данным Go.' }])
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [alternatives, setAlternatives] = useState([])
  const [proposal, setProposal] = useState(null)
  const [cartResult, setCartResult] = useState(null)
  const [chatOpen, setChatOpen] = useState(true)

  const selected = selectedProduct || products.find((item) => item.id === selectedId) || null
  const visibleProducts = useMemo(() => products, [products])

  async function searchProducts(value) {
    setQuery(value); setError(''); setLoading(true)
    if (demoMode) { setProducts(demoProducts.filter((item) => !value || `${item.id} ${item.article} ${item.name}`.toLowerCase().includes(value.toLowerCase()))); setLoading(false); return }
    try { const result = await api.search(value); setProducts((result?.items || []).map(normalizeProduct)); } catch (requestError) { setError(`Поиск недоступен: ${requestError.message}`) } finally { setLoading(false) }
  }

  async function openProduct(id) {
    setSelectedId(id); setError(''); setLoading(true); setAlternatives([])
    if (demoMode) { setSelectedProduct(demoProducts.find((item) => item.id === id) || null); setLoading(false); return }
    try { setSelectedProduct(normalizeProduct(await api.product(id))) } catch (requestError) { setSelectedProduct(null); setError(`Карточку не удалось загрузить: ${requestError.message}`) } finally { setLoading(false) }
  }

  async function sendMessage(text = message) {
    const cleanText = text.trim(); if (!cleanText || loading) return
    setMessages((current) => [...current, { from: 'user', text: cleanText }]); setMessage(''); setError(''); setLoading(true)
    if (demoMode) { respondInDemo(cleanText); setLoading(false); return }
    try {
      const result = await api.chat({ session_id: 'browser-session', message: cleanText, selected_product_id: selected?.id || selectedId || undefined })
      setMessages((current) => [...current, { from: 'assistant', text: result?.reply || 'Сервер не вернул текст ответа.' }])
      const returnedProducts = (result?.products || []).map(normalizeProduct)
      if (returnedProducts.length) { setProducts(returnedProducts); setSelectedProduct(returnedProducts[0]); setSelectedId(returnedProducts[0].id) }
      setProposal(result?.pending_action || null)
    } catch (requestError) { setError(`Ассистент недоступен: ${requestError.message}`); setMessages((current) => [...current, { from: 'assistant', text: 'Не удалось получить ответ от Go. Данные не подставлены.' }]) } finally { setLoading(false) }
  }

  function respondInDemo(text) {
    const lower = text.toLowerCase(); let reply = 'В демо-выборке нет подтверждённого ответа на этот вопрос.'
    if (lower.includes('515291') || lower.includes('есть')) { setSelectedProduct(demoProducts[0]); setSelectedId('515291'); reply = 'ID 515291 найден. Цена и остаток показаны из демо-данных. Номинальный ток требует уточнения: 160 А против 250 А.' }
    if (lower.includes('замен')) { setSelectedProduct(demoProducts[1]); setSelectedId('48783'); setAlternatives([{ product: demoProducts[1], reason: 'Совпадают полюса, напряжение и отключающая способность.', differences: ['Номинальный ток отличается'] }]); reply = 'Проверенный в демо-паре вариант найден. Перед монтажом нужна проверка специалиста.' }
    if (lower.includes('куп') || lower.includes('добав')) { setProposal({ product_id: selected?.id || '515291', quantity: 2, confirmation_token: 'demo-token', product: selected || demoProducts[0] }); reply = 'Сформировала предложение на 2 шт. Корзина не изменится без отдельного подтверждения.' }
    setMessages((current) => [...current, { from: 'assistant', text: reply }])
  }

  async function loadAlternatives() {
    setError(''); setLoading(true)
    if (demoMode) { setAlternatives([{ product: demoProducts[1], reason: 'Совпадают полюса, напряжение и отключающая способность.', differences: ['Номинальный ток отличается'] }]); setLoading(false); return }
    try { const result = await api.alternatives(selected.id); setAlternatives(result?.items || []) } catch (requestError) { setError(`Аналоги не удалось загрузить: ${requestError.message}`) } finally { setLoading(false) }
  }

  async function confirmAdd() {
    if (!proposal || loading) return
    setLoading(true); setError(''); setCartResult(null)
    if (demoMode) { setCartResult({ status: 'added', mode: 'demo' }); setProposal(null); setLoading(false); return }
    try { const result = await api.confirm(proposal.confirmation_token); setCartResult(result); setProposal(null) } catch (requestError) { if (requestError.status === 409) setError('Данные изменились. Проверьте новые цену и остаток и подтвердите предложение заново.'); else setError(`Добавление не выполнено: ${requestError.message}`) } finally { setLoading(false) }
  }

  function toggleMode() {
    const next = !demoMode; setDemoMode(next); setError(''); setAlternatives([]); setProposal(null); setCartResult(null)
    if (next) { setProducts(demoProducts); setSelectedProduct(demoProducts[0]); setSelectedId('515291'); setMessages([{ from: 'assistant', text: 'Демо-режим включён. Это локальные данные, не актуальная корзина EKT.' }]) }
    else { setProducts([]); setSelectedProduct(null); setSelectedId(null); setMessages([{ from: 'assistant', text: 'API-режим включён. Задайте вопрос, чтобы обратиться к Go.' }]) }
  }

  return <div className={`app-shell ${chatOpen ? 'chat-active' : ''}`}>
    <div className="utility-bar"><div>⌖ Астана</div><div className="utility-links"><span>Личный кабинет</span><span>B2B · EKT PRO</span><span>Покупателям⌄</span><span>Оставить заявку</span><strong>+7 (700) 222-05-14</strong></div></div>
    <header className="main-header"><div className="brand"><span>ГРУППА КОМПАНИЙ</span><b>ЭЛЕКТРОКОМПЛЕКТ</b></div><button className="catalog-button">Каталог <span>☷</span></button><label className="global-search"><span>⌕</span><input value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && searchProducts(query)} placeholder="Поиск по названию, ID или артикулу" /><kbd>⌘ K</kbd></label><div className="header-actions"><button className={demoMode ? 'mode-chip demo' : 'mode-chip'} onClick={toggleMode}>{demoMode ? 'Демо-режим' : 'API-режим'}</button><button onClick={() => setChatOpen(true)}>▱ Корзина</button></div></header>
    <main><section className="welcome-row"><div><p className="eyebrow">EKT SMART SEARCH <span>{demoMode ? 'ДЕМО-РЕЖИМ' : 'GO API'}</span></p><h1>Подберём нужное.<br /><em>Проверим каждую деталь.</em></h1><p className="intro">Центральный сценарий проходит через ИИ-помощника. Карточки, аналоги и корзина используют данные из одного источника.</p></div><div className="trust-panel"><div className="pulse-dot"></div><div><b>{demoMode ? 'Локальная демонстрация' : 'Источник: Go API'}</b><span>{demoMode ? 'Не актуальная корзина EKT' : 'Ключи остаются на сервере'}</span></div><small>›</small></div></section><div className="search-wrap"><span className="search-icon">⌕</span><input value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && searchProducts(query)} placeholder="Например, 515291 или 200300285_" /><button onClick={() => searchProducts(query)}>Найти <span>↗</span></button></div><div className="quick-row"><span>Начните с чата:</span>{suggestions.map((item) => <button key={item} onClick={() => sendMessage(item)}>{item}</button>)}</div>
      {error && <div className="api-error"><b>Ошибка</b><span>{error}</span>{!demoMode && <button onClick={toggleMode}>Открыть демо-режим</button>}</div>}
      <div className="content-grid"><section className="catalog-column"><div className="section-heading"><div><span className="section-kicker">{demoMode ? 'ДЕМО-КАТАЛОГ' : 'КАТАЛОГ ИЗ GO'} · {visibleProducts.length} ПОЗИЦИИ</span><h2>Товары для вашего проекта</h2></div><div className="view-toggle"><button className="active">▦</button><button>☷</button></div></div>{loading && <div className="loading-state">Проверяю данные…</div>}{!loading && !visibleProducts.length && <div className="empty-state"><span>⌕</span><h3>Каталог пока не загружен</h3><p>Задайте вопрос в чате или проверьте, запущен ли Go API.</p></div>}<div className="product-list">{visibleProducts.map((product) => <article className={`product-card ${selectedId === product.id ? 'selected' : ''}`} key={product.id} onClick={() => openProduct(product.id)}><div className="product-image"><img src={product.image} alt="" /><span className={product.quantity === null ? 'unknown-stock' : product.quantity ? 'in-stock' : 'out-stock'}>{product.quantity === null ? 'Наличие неизвестно' : product.quantity ? 'В наличии' : 'Нет в наличии'}</span></div><div className="product-info"><span className="product-type">{product.type || 'Товар EKT'}</span><h3>{product.name}</h3><div className="meta"><span>ID {product.id}</span><span>Арт. {product.article || 'Нет данных'}</span></div><div className="product-bottom"><b>{priceLabel(product.price)}</b><span className={product.quantity === null ? 'muted' : product.quantity ? 'stock' : 'muted'}>{product.quantity === null ? 'Наличие неизвестно' : product.quantity ? `${product.quantity} шт. на складах` : 'Остаток 0 шт.'}</span></div></div><button className="card-arrow" aria-label="Открыть товар">↗</button></article>)}</div></section>
        <aside className="detail-column"><div className="detail-top"><span className="section-kicker">ПОДРОБНАЯ КАРТОЧКА</span><div className="detail-actions"><button>♡</button><button>⋯</button></div></div>{selected ? <><div className="detail-image"><img src={selected.image} alt="" /><span className="verified">✓ {demoMode ? 'Демо-данные' : 'Получено из Go'}</span></div><span className="product-type">{selected.type || 'Товар EKT'}</span><h2>{selected.name}</h2><div className="detail-meta"><span>Бренд <b>{selected.brand || 'Нет данных'}</b></span><span>Артикул <b>{selected.article || 'Нет данных'}</b></span></div><div className="price-row"><div><small>Цена</small><strong>{priceLabel(selected.price)}</strong></div><div className={selected.quantity === null ? 'availability unknown' : selected.quantity ? 'availability' : 'availability unavailable'}><i></i>{selected.quantity === null ? 'Наличие неизвестно' : selected.quantity ? `В наличии · ${selected.quantity} шт.` : 'Нет в наличии'}</div></div>{(selected.data_issues || selected.warning) && <div className="warning"><span>!</span><p><b>Нужно внимание</b>{(selected.data_issues || [selected.warning]).join(' ')}</p></div>}<div className="characteristics"><div className="subheading"><b>Характеристики</b><span>Только из источника ↗</span></div>{(selected.characteristics || []).map(([key, value]) => <div className="characteristic" key={key}><span>{key}</span><b className={String(value).toLowerCase().includes('уточ') ? 'needs-check' : ''}>{value}</b></div>)}</div>{selected.quantity === 0 && <button className="outline-button" onClick={loadAlternatives}>Подобрать замену <span>→</span></button>}<a className="ekt-link" href={selected.url || 'https://ekt.kz'} target="_blank" rel="noreferrer">Открыть на ekt.kz ↗</a></> : <div className="empty-state"><span>⌕</span><h3>Выберите товар</h3><p>Карточка будет загружена из Go API.</p></div>}</aside></div>
      {alternatives.length > 0 && <section className="compare-section"><div className="section-heading"><div><span className="section-kicker">ПРОВЕРЕННЫЙ ПОДБОР</span><h2>Возможная замена</h2></div></div>{alternatives.map((item) => <div className="compare-card" key={item.product.id}><div className="compare-product"><img src={item.product.image} alt="" /><div><span className="product-type">{item.product.type}</span><h3>{item.product.name}</h3><span className="stock">{item.product.quantity > 0 ? `✓ В наличии · ${item.product.quantity} шт.` : 'Наличие неизвестно'}</span></div></div><div className="match-score"><strong>Проверено</strong><span>{item.reason}</span></div><div className="compare-note"><b>Отличия</b><p>{(item.differences || []).join('; ') || 'Нет данных об отличиях.'}</p></div></div>)}</section>}
      {cartResult && <div className="success-state"><b>{cartResult.mode === 'demo' ? 'Демо-корзина' : 'Добавлено'}</b><span>{cartResult.status || 'Позиция добавлена после проверки.'}</span>{cartResult.cart_url && <a href={cartResult.cart_url} target="_blank" rel="noreferrer">Открыть корзину ↗</a>}</div>}
    </main>
    <button className={`chat-launcher ${chatOpen ? 'hidden' : ''}`} onClick={() => setChatOpen(true)}>✦ <span>Спросить EKT AI</span></button>
    {chatOpen && <aside className="chat-panel"><div className="chat-header"><div className="assistant-avatar">✦</div><div><b>EKT AI</b><span>{demoMode ? 'Демо-режим' : 'Ассистент через Go API'} <i></i></span></div><button onClick={() => setChatOpen(false)}>×</button></div><div className="chat-context"><span>◉</span> {demoMode ? 'Локальные данные · не актуальная корзина' : 'Ответы только из подтверждённых данных Go'}</div><div className="messages">{messages.map((item, index) => <div className={`message ${item.from}`} key={`${item.text}-${index}`}><p>{item.text}</p></div>)}{proposal && <div className="proposal"><span className="proposal-label">ПОДТВЕРЖДЕНИЕ</span><b>ID {proposal.product_id} · {proposal.quantity} шт.</b><span>Ничего не добавлено без подтверждения</span><button onClick={confirmAdd} disabled={loading}>Подтвердить добавление <span>→</span></button><button className="cancel-proposal" onClick={() => setProposal(null)}>Отмена</button></div>}<div className="chat-suggestions">{suggestions.map((item) => <button key={item} onClick={() => sendMessage(item)}>{item}</button>)}</div></div><form className="chat-input" onSubmit={(event) => { event.preventDefault(); sendMessage() }}><input value={message} onChange={(event) => setMessage(event.target.value)} placeholder="Напишите вопрос..." /><button aria-label="Отправить" disabled={loading}>↑</button></form><div className="chat-footnote">{demoMode ? 'Демо-корзина · локальные данные' : 'POST /api/chat · POST /api/cart/confirm'}</div></aside>}
  </div>
}

export default App
