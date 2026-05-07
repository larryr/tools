(() => {
  const palette = JSON.parse(document.getElementById("palette").textContent || "{}");
  const memberColor = Object.fromEntries((palette.members || []).map(m => [m.id, m.color]));
  const labelColor = Object.fromEntries((palette.labels || []).map(l => [l.name, l.color]));

  // Apply colors looked up from the palette.
  document.querySelectorAll(".chip[data-member]").forEach(el => {
    const c = memberColor[el.dataset.member];
    if (c) el.style.background = c;
  });
  document.querySelectorAll(".pill[data-label]").forEach(el => {
    const c = labelColor[el.dataset.label];
    if (c) el.style.background = c;
  });

  let dragging = null;

  function onDragStart(e) {
    dragging = e.currentTarget;
    dragging.classList.add("dragging");
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", dragging.dataset.id);
  }

  function onDragEnd() {
    if (dragging) dragging.classList.remove("dragging");
    document.querySelectorAll(".drop-marker").forEach(m => m.remove());
    document.querySelectorAll(".cards.drag-over").forEach(c => c.classList.remove("drag-over"));
    dragging = null;
  }

  function insertionIndex(ul, y) {
    const cards = [...ul.querySelectorAll(".card:not(.dragging)")];
    for (let i = 0; i < cards.length; i++) {
      const r = cards[i].getBoundingClientRect();
      if (y < r.top + r.height / 2) return i;
    }
    return cards.length;
  }

  function onDragOver(e) {
    if (!dragging) return;
    e.preventDefault();
    const ul = e.currentTarget;
    ul.classList.add("drag-over");
    const idx = insertionIndex(ul, e.clientY);
    const cards = [...ul.querySelectorAll(".card:not(.dragging)")];
    document.querySelectorAll(".drop-marker").forEach(m => m.remove());
    const marker = document.createElement("li");
    marker.className = "drop-marker";
    if (idx >= cards.length) ul.appendChild(marker);
    else ul.insertBefore(marker, cards[idx]);
  }

  function onDragLeave(e) {
    if (e.currentTarget.contains(e.relatedTarget)) return;
    e.currentTarget.classList.remove("drag-over");
  }

  async function onDrop(e) {
    if (!dragging) return;
    e.preventDefault();
    const ul = e.currentTarget;
    const id = dragging.dataset.id;
    const idx = insertionIndex(ul, e.clientY);
    // Optimistic DOM update.
    const cards = [...ul.querySelectorAll(".card:not(.dragging)")];
    if (idx >= cards.length) ul.appendChild(dragging);
    else ul.insertBefore(dragging, cards[idx]);
    onDragEnd();
    try {
      const res = await fetch(`/api/cards/${encodeURIComponent(id)}/move`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ column: ul.dataset.column, index: idx }),
      });
      if (!res.ok) throw new Error(await res.text());
      updateWipCounts();
    } catch (err) {
      console.error("move failed:", err);
      location.reload();
    }
  }

  function updateWipCounts() {
    document.querySelectorAll(".column").forEach(col => {
      const head = col.querySelector(".wip");
      if (!head) return;
      const count = col.querySelectorAll(".card").length;
      const txt = head.textContent.trim();
      const slash = txt.indexOf("/");
      head.textContent = slash >= 0 ? `${count}${txt.slice(slash)}` : `${count}`;
    });
  }

  function bindCard(card) {
    card.addEventListener("dragstart", onDragStart);
    card.addEventListener("dragend", onDragEnd);
    card.addEventListener("dblclick", onCardEdit);
  }

  async function onCardEdit(e) {
    const card = e.currentTarget;
    const titleEl = card.querySelector(".title");
    const old = titleEl.textContent;
    const next = prompt("Edit card title:", old);
    if (next == null || next.trim() === "" || next === old) return;
    const id = card.dataset.id;
    titleEl.textContent = next;
    const res = await fetch(`/api/cards/${encodeURIComponent(id)}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: next }),
    });
    if (!res.ok) {
      titleEl.textContent = old;
      alert("edit failed: " + (await res.text()));
    }
  }

  async function onAddCard(e) {
    const col = e.currentTarget.dataset.column;
    const title = prompt("New card title:");
    if (!title || !title.trim()) return;
    const res = await fetch("/api/cards", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ column: col, title: title.trim() }),
    });
    if (!res.ok) {
      alert("add failed: " + (await res.text()));
      return;
    }
    location.reload();
  }

  document.querySelectorAll(".card").forEach(bindCard);
  document.querySelectorAll(".cards").forEach(ul => {
    ul.addEventListener("dragover", onDragOver);
    ul.addEventListener("dragleave", onDragLeave);
    ul.addEventListener("drop", onDrop);
  });
  document.querySelectorAll(".add-card").forEach(b => b.addEventListener("click", onAddCard));
})();
