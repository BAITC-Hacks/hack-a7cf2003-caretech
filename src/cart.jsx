import { useState } from 'react'
import { createRoot } from 'react-dom/client'
import './cart.css'

export function Cart() {
  const [items, setItems] = useState(() => JSON.parse(localStorage.getItem('ekt-cart') || '[]'))
  function removeItem(id) {
    const next = items.filter((item) => item.id !== id)
    setItems(next)
    localStorage.setItem('ekt-cart', JSON.stringify(next))
  }
  return <div className="saved-cart"><header className="saved-header"><a href="index.html" className="saved-brand">ЭЛЕКТРОКОМПЛЕКТ</a><a href="catalog.html" className="saved-back">← Вернуться в каталог</a></header><main className="saved-main"><p className="saved-kicker">ВАШИ ТОВАРЫ</p><h1>Сохранённые товары</h1><p className="saved-intro">Здесь хранятся товары, которые вы отметили для проекта. Список сохраняется в этом браузере.</p>{items.length === 0 ? <section className="saved-empty"><span>♡</span><h2>Пока ничего нет</h2><p>Откройте товар в каталоге и нажмите на сердечко, чтобы сохранить его здесь.</p><a href="catalog.html">Открыть каталог</a></section> : <section className="saved-list">{items.map((item) => <article className="saved-item" key={item.id}><img src={item.image || '/product-placeholder.svg'} alt="" /><div><span>{item.type || 'Товар EKT'}</span><h2>{item.name}</h2><p>ID {item.id} · Арт. {item.article || 'Нет данных'}</p><strong>{item.price || 'Цена не получена'}</strong></div><button onClick={() => removeItem(item.id)} aria-label={`Удалить ${item.name}`}>×</button></article>)}</section>}</main></div>
}

createRoot(document.getElementById('cart-root')).render(<Cart />)
