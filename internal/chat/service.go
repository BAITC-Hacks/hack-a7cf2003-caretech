package chat

import (
	"fmt"
	"strings"
	"sync"

	"hackalem-ekt-backend/internal/cart"
	"hackalem-ekt-backend/internal/catalog"
)

type Request struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type Response struct {
	Reply      string             `json:"reply"`
	Products   []catalog.Product  `json:"products,omitempty"`
	PendingAdd *PendingAdd        `json:"pending_add,omitempty"`
	CartURL    string             `json:"cart_url,omitempty"`
}

type PendingAdd struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type session struct {
	pending *PendingAdd
}

type Service struct {
	catalog  catalog.Repository
	carts    cart.Store
	mu       sync.Mutex
	sessions map[string]*session
}

func NewService(products catalog.Repository, carts cart.Store) *Service {
	return &Service{catalog: products, carts: carts, sessions: make(map[string]*session)}
}

func (service *Service) Handle(request Request) Response {
	service.mu.Lock()
	defer service.mu.Unlock()

	current := service.getSession(request.SessionID)
	message := strings.ToLower(strings.TrimSpace(request.Message))
	if isConfirmation(message) && current.pending != nil {
		pending := *current.pending
		if err := service.carts.Add(request.SessionID, pending.SKU, pending.Quantity); err != nil {
			return Response{Reply: err.Error()}
		}
		current.pending = nil
		return Response{Reply: "Готово, товар добавлен в корзину.", CartURL: "/api/cart?session_id=" + request.SessionID}
	}

	query := normalizeQuery(message)
	products := service.catalog.Search(query)
	if len(products) > 0 {
		product := products[0]
		if product.Stock == 0 {
			return Response{Reply: "Товар временно отсутствует. Подобрал аналог с тем же током и характеристикой.", Products: service.catalog.Alternatives(product)}
		}
		current.pending = &PendingAdd{SKU: product.SKU, Quantity: 1}
		return Response{
			Reply:      fmt.Sprintf("%s в наличии: %d шт., цена %d тг. Добавить 1 шт. в корзину?", product.Name, product.Stock, product.Price),
			Products:   []catalog.Product{product},
			PendingAdd: current.pending,
		}
	}

	if strings.Contains(message, "достав") || strings.Contains(message, "оплат") || strings.Contains(message, "услов") {
		return Response{Reply: "Оплата: безналичный расчет и картой на сайте. Доставка рассчитывается по городу и объему заказа. Минимальная партия зависит от позиции; точные условия подтвердит менеджер."}
	}
	return Response{Reply: "Уточните артикул или название товара. Я проверю наличие, характеристики, сертификат и аналоги."}
}

func normalizeQuery(message string) string {
	fields := strings.FieldsFunc(message, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == ' ')
	})
	filtered := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "есть", "нужен", "нужна", "нужно", "хочу", "покажи", "подскажи", "товар", "поиск", "ищу", "можно", "какой", "какая":
			continue
		default:
			filtered = append(filtered, field)
		}
	}
	return strings.TrimSpace(strings.Join(filtered, " "))
}

func (service *Service) Confirm(sessionID string) (string, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	current := service.getSession(sessionID)
	if current.pending == nil {
		return "", fmt.Errorf("нет ожидающего добавления в корзину")
	}
	if err := service.carts.Add(sessionID, current.pending.SKU, current.pending.Quantity); err != nil {
		return "", err
	}
	current.pending = nil
	return "/api/cart?session_id=" + sessionID, nil
}

func (service *Service) getSession(sessionID string) *session {
	current := service.sessions[sessionID]
	if current == nil {
		current = &session{}
		service.sessions[sessionID] = current
	}
	return current
}

func isConfirmation(message string) bool {
	for _, word := range []string{"да", "добавь", "подтверждаю", "confirm"} {
		if strings.Contains(message, word) {
			return true
		}
	}
	return false
}