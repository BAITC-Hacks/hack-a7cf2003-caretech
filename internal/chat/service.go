package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"hackalem-ekt-backend/internal/cart"
	"hackalem-ekt-backend/internal/catalog"
	"hackalem-ekt-backend/internal/llm"
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
	llm      llm.Client
	mu       sync.Mutex
	sessions map[string]*session
}

type Option func(*Service)

func WithLLMClient(client llm.Client) Option {
	return func(service *Service) {
		service.llm = client
	}
}

func NewService(products catalog.Repository, carts cart.Store, opts ...Option) *Service {
	service := &Service{catalog: products, carts: carts, sessions: make(map[string]*session)}
	for _, opt := range opts {
		if opt != nil {
			opt(service)
		}
	}
	return service
}

func (service *Service) Handle(request Request) Response {
	service.mu.Lock()
	defer service.mu.Unlock()

	current := service.getSession(request.SessionID)
	message := strings.ToLower(strings.TrimSpace(request.Message))
	if isConfirmation(message) {
		if current.pending == nil {
			return Response{Reply: "Нет ожидающего добавления в корзину. Сначала уточните товар и подтвердите его добавление."}
		}
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
	if service.llm != nil {
		if reply, ok := service.generateLLMReply(request); ok {
			return Response{Reply: reply}
		}
	}
	return Response{Reply: "Уточните артикул или название товара. Я проверю наличие, характеристики, сертификат и аналоги."}
}

func (service *Service) generateLLMReply(request Request) (string, bool) {
	if service.llm == nil {
		return "", false
	}
	ctx := context.Background()
	prompt := fmt.Sprintf("Ты помощник по каталогу EKT. Ответь на вопрос клиента кратко и по делу. Используй только данные каталога и оферты, а не общие предположения.\n\nВопрос клиента: %s\n\nЕсли это неуловимый вопрос про товар, уточни артикул или тип оборудования, но не придумывай остатки и цену.", request.Message)
	reply, err := service.llm.Generate(ctx, "Ты консультант EKT по электрооборудованию. Отвечай сухо, по делу, без выдумок, и всегда опирайся на каталог и доступные данные.", prompt)
	if err != nil || strings.TrimSpace(reply) == "" {
		return "", false
	}
	return reply, true
}

func normalizeQuery(message string) string {
	message = strings.ToLower(strings.TrimSpace(message))
	tokens := strings.Fields(message)
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.Trim(token, "\t\n\r.,!?;:()[]{}\"'`<>/\\")
		if token == "" {
			continue
		}
		switch token {
		case "есть", "нужен", "нужна", "нужно", "хочу", "покажи", "подскажи", "товар", "поиск", "ищу", "можно", "какой", "какая":
			continue
		default:
			filtered = append(filtered, token)
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