package lkan

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// MoveCard moves the card with id to the column with toColumn at the given
// index. Negative or too-large indices are clamped. If the card is already
// in toColumn, indices are interpreted against the column with the card
// removed first, then it is reinserted.
func MoveCard(b *Board, id, toColumn string, index int) error {
	dst := b.FindColumn(toColumn)
	if dst == nil {
		return fmt.Errorf("%w: column %q", ErrNotFound, toColumn)
	}
	_, src, pos := b.FindCard(id)
	if src == nil {
		return fmt.Errorf("%w: card %q", ErrNotFound, id)
	}
	card := src.Cards[pos]
	src.Cards = append(src.Cards[:pos], src.Cards[pos+1:]...)
	if index < 0 {
		index = 0
	}
	if index > len(dst.Cards) {
		index = len(dst.Cards)
	}
	dst.Cards = append(dst.Cards, Card{})
	copy(dst.Cards[index+1:], dst.Cards[index:])
	dst.Cards[index] = card
	return nil
}

// AddCard appends a new card to the named column and returns its id.
func AddCard(b *Board, columnID string, c Card) (string, error) {
	col := b.FindColumn(columnID)
	if col == nil {
		return "", fmt.Errorf("%w: column %q", ErrNotFound, columnID)
	}
	if strings.TrimSpace(c.Title) == "" {
		return "", fmt.Errorf("title required")
	}
	if c.ID == "" {
		c.ID = newID()
	}
	col.Cards = append(col.Cards, c)
	return c.ID, nil
}

// CardPatch holds optional fields for EditCard. A nil pointer leaves the
// field unchanged.
type CardPatch struct {
	Title       *string
	Description *string
	Assignees   *[]string
	Labels      *[]string
	Due         *string
}

// EditCard applies a partial update to the card with id.
func EditCard(b *Board, id string, p CardPatch) error {
	card, _, _ := b.FindCard(id)
	if card == nil {
		return fmt.Errorf("%w: card %q", ErrNotFound, id)
	}
	if p.Title != nil {
		card.Title = *p.Title
	}
	if p.Description != nil {
		card.Description = *p.Description
	}
	if p.Assignees != nil {
		card.Assignees = append([]string(nil), (*p.Assignees)...)
	}
	if p.Labels != nil {
		card.Labels = append([]string(nil), (*p.Labels)...)
	}
	if p.Due != nil {
		card.Due = *p.Due
	}
	return nil
}

// DeleteCard removes the card with id.
func DeleteCard(b *Board, id string) error {
	_, col, pos := b.FindCard(id)
	if col == nil {
		return fmt.Errorf("%w: card %q", ErrNotFound, id)
	}
	col.Cards = append(col.Cards[:pos], col.Cards[pos+1:]...)
	return nil
}

func newID() string {
	var buf [6]byte
	_, _ = rand.Read(buf[:])
	return "c-" + hex.EncodeToString(buf[:])
}
