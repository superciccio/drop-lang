package ui

const DefaultCSS = `
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background: #0a0a0a;
  color: #ededed;
  line-height: 1.6;
  padding: 2rem 1rem;
}

.container {
  max-width: 720px;
  margin: 0 auto;
}

h1 {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  color: #fff;
}

/* Tables (auto-render lists) */
table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 1.5rem;
}

th {
  text-align: left;
  padding: 0.6rem 0.8rem;
  border-bottom: 2px solid #333;
  font-weight: 600;
  color: #999;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

td {
  padding: 0.6rem 0.8rem;
  border-bottom: 1px solid #1a1a1a;
}

tr:nth-child(even) td { background: #111; }
tr:hover td { background: #1a1a1a; }

/* Cards (auto-render objects) */
.card {
  background: #111;
  border: 1px solid #222;
  border-radius: 8px;
  padding: 1.2rem;
  margin-bottom: 1rem;
}

.card-row {
  display: flex;
  justify-content: space-between;
  padding: 0.4rem 0;
  border-bottom: 1px solid #1a1a1a;
}

.card-row:last-child { border-bottom: none; }

.card-key {
  color: #888;
  font-size: 0.85rem;
}

.card-value { color: #ededed; }

/* Text blocks */
.text-block {
  margin-bottom: 1rem;
  line-height: 1.8;
}

/* Rows */
.row {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.6rem 0;
  border-bottom: 1px solid #1a1a1a;
}

.row:last-child { border-bottom: none; }

/* Buttons */
button, .btn {
  background: #ededed;
  color: #0a0a0a;
  border: none;
  padding: 0.4rem 1rem;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

button:hover, .btn:hover { background: #fff; }

button.danger { background: #dc3545; color: #fff; }
button.danger:hover { background: #e04555; }

/* Forms */
.form-section {
  background: #111;
  border: 1px solid #222;
  border-radius: 8px;
  padding: 1.2rem;
  margin-top: 1.5rem;
}

.form-title {
  font-size: 0.9rem;
  color: #888;
  margin-bottom: 1rem;
  font-weight: 500;
}

input[type="text"] {
  width: 100%;
  padding: 0.5rem 0.8rem;
  background: #0a0a0a;
  border: 1px solid #333;
  border-radius: 6px;
  color: #ededed;
  font-size: 0.9rem;
  margin-bottom: 0.8rem;
  outline: none;
  transition: border-color 0.15s;
}

input[type="text"]:focus { border-color: #ededed; }

input[type="text"]::placeholder { color: #555; }

/* Links */
a {
  color: #8bb4ff;
  text-decoration: none;
}

a:hover { text-decoration: underline; }

/* Images */
img {
  max-width: 100%;
  border-radius: 8px;
  margin-bottom: 1rem;
}

/* Responsive */
@media (max-width: 600px) {
  .row { flex-direction: column; align-items: flex-start; gap: 0.4rem; }
  table { font-size: 0.85rem; }
}
`
