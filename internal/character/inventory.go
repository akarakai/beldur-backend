package character

const ItemDefaultCapacity = 10

type Item struct{}

type Inventory struct {
	capacity int
	items    []Item
}

func NewEmptyInventory() Inventory {
	return Inventory{
		capacity: ItemDefaultCapacity,
		items:    make([]Item, 0, ItemDefaultCapacity),
	}
}

func (i *Inventory) AddItem(item Item) {
	i.items = append(i.items, item)
}
