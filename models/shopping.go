package models

type ShoppingItem struct {
	ID        int     `json:"id,omitempty"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Purchased bool    `json:"purchased,omitempty"`
}

type UpdateShoppingItem struct {
	Name      *string  `json:"name"`
	Category  *string  `json:"category"`
	Price     *float64 `json:"price"`
	Quantity  *int     `json:"quantity"`
	Purchased *bool    `json:"purchased,omitempty"`
}

type APIStatus struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ShopList struct {
	List map[int]*ShoppingItem `json:"list"`
}

func NewShopList() *ShopList {
	return &ShopList{
		make(map[int]*ShoppingItem),
	}
}

var (
	Cart   = NewShopList()
	NextID = 1
)

func (sl *ShopList) AddItem(item ShoppingItem) {
	item.ID = NextID
	NextID++
	sl.List[item.ID] = &item
}

func (sl *ShopList) RemoveItemById(id int) {
	delete(sl.List, id)
}

func (sl *ShopList) GetAllItems() map[int]*ShoppingItem {
	return sl.List
}

func (sl *ShopList) GetItemById(id int) *ShoppingItem {
	return sl.List[id]
}

func (sl *ShopList) FullUpdateItem(id int, item ShoppingItem) {
	sl.List[id] = &item
}

func (sl *ShopList) GetItemsByCategory(category string) []*ShoppingItem {
	var sortedItems []*ShoppingItem
	for _, v := range sl.List {
		if v.Category == category {
			sortedItems = append(sortedItems, v)
		}
	}
	return sortedItems
}
