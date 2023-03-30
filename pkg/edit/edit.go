package edit

import "github.com/lukasschwab/tiir/pkg/text"

type Editor interface {
	Update(initial *text.Text) (final *text.Text, err error)
}
