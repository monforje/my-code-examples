(function() {
  const DEFAULT_OPTIONS = {
    gap: 8,
    stepDelay: 18,
    widthDuration: 55,
    heightDuration: 65,
  };

  function wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  function nextFrame() {
    return new Promise(resolve => requestAnimationFrame(resolve));
  }

  function getColumns(width) {
    if (width < 560) return 4;
    if (width < 900) return 6;
    if (width < 1180) return 8;
    return 12;
  }

  function uniqueShapes(shapes, columns) {
    const seen = {};
    return shapes.filter(([w, h]) => {
      const width = Math.min(Math.max(w, 2), columns);
      const key = width + 'x' + h;
      if (seen[key]) return false;
      seen[key] = true;
      return width > 0 && h > 0;
    }).map(([w, h]) => [Math.min(Math.max(w, 2), columns), h]);
  }

  function getShapeOptions(item, index, columns) {
    const mobileShapes = [[4, 2], [2, 2], [2, 3], [4, 2], [2, 2], [2, 3]];
    const smallShapes = [[3, 3], [3, 2], [3, 3], [2, 3], [4, 2], [3, 3]];
    const tabletShapes = [[4, 3], [4, 2], [4, 3], [3, 3], [4, 2], [3, 3]];
    const desktopShapes = [[4, 3], [3, 2], [5, 2], [2, 3], [4, 2], [3, 3]];
    const shapes = columns <= 4 ? mobileShapes : columns <= 6 ? smallShapes : columns <= 8 ? tabletShapes : desktopShapes;
    const primary = shapes[index % shapes.length];
    const contentWeight = String(item.title || '').length + Math.round(String(item.description || '').length * 0.35) + (item.tags || []).length * 8;
    const needsRoom = item.complexity === 'Senior' || item.complexity === 'Hard' || contentWeight > 86;

    return uniqueShapes([
      primary,
      needsRoom ? [primary[0] + 1, primary[1]] : [primary[0], primary[1] + 1],
      [primary[0] - 1, primary[1] + 1],
    ], columns);
  }

  function canPlace(usedRows, columns, col, row, colSpan, rowSpan) {
    if (col + colSpan > columns) return false;

    for (let r = row; r < row + rowSpan; r++) {
      for (let c = col; c < col + colSpan; c++) {
        if (usedRows[r] && usedRows[r][c]) return false;
      }
    }

    return true;
  }

  function markPlace(usedRows, col, row, colSpan, rowSpan) {
    for (let r = row; r < row + rowSpan; r++) {
      if (!usedRows[r]) usedRows[r] = [];
      for (let c = col; c < col + colSpan; c++) {
        usedRows[r][c] = true;
      }
    }
  }

  function touchLength(a, b) {
    let touch = 0;

    if (a.col + a.colSpan === b.col || b.col + b.colSpan === a.col) {
      touch += Math.max(0, Math.min(a.row + a.rowSpan, b.row + b.rowSpan) - Math.max(a.row, b.row));
    }

    if (a.row + a.rowSpan === b.row || b.row + b.rowSpan === a.row) {
      touch += Math.max(0, Math.min(a.col + a.colSpan, b.col + b.colSpan) - Math.max(a.col, b.col));
    }

    return touch;
  }

  function rectDistance(a, b) {
    const dx = a.col + a.colSpan < b.col
      ? b.col - (a.col + a.colSpan)
      : b.col + b.colSpan < a.col
        ? a.col - (b.col + b.colSpan)
        : 0;
    const dy = a.row + a.rowSpan < b.row
      ? b.row - (a.row + a.rowSpan)
      : b.row + b.rowSpan < a.row
        ? a.row - (b.row + b.rowSpan)
        : 0;

    return dx + dy;
  }

  function adjacencyScore(usedRows, candidate) {
    let score = 0;
    const left = candidate.col - 1;
    const right = candidate.col + candidate.colSpan;
    const top = candidate.row - 1;
    const bottom = candidate.row + candidate.rowSpan;

    for (let r = candidate.row; r < candidate.row + candidate.rowSpan; r++) {
      if (usedRows[r] && usedRows[r][left]) score++;
      if (usedRows[r] && usedRows[r][right]) score++;
    }

    for (let c = candidate.col; c < candidate.col + candidate.colSpan; c++) {
      if (usedRows[top] && usedRows[top][c]) score++;
      if (usedRows[bottom] && usedRows[bottom][c]) score++;
    }

    return score;
  }

  function alignmentScore(candidate, placed) {
    let score = 0;
    placed.forEach(rect => {
      if (candidate.col === rect.col) score++;
      if (candidate.row === rect.row) score++;
      if (candidate.col + candidate.colSpan === rect.col + rect.colSpan) score++;
      if (candidate.row + candidate.rowSpan === rect.row + rect.rowSpan) score++;
    });
    return score;
  }

  function horizontalScore(candidate, placed) {
    let score = 0;
    const last = placed[placed.length - 1];

    if (last && candidate.row === last.row && candidate.col === last.col + last.colSpan) score += 12;
    if (last && candidate.row === last.row && candidate.col > last.col) score += 6;

    placed.forEach(rect => {
      if (candidate.row === rect.row && candidate.col === rect.col + rect.colSpan) score += 5;
      if (candidate.row === rect.row && candidate.col + candidate.colSpan === rect.col) score += 3;
    });

    return score;
  }

  function findBestPlace(usedRows, columns, colSpan, rowSpan, placed, maxRow) {
    let best = null;
    const last = placed[placed.length - 1];
    const rowLimit = Math.max(maxRow + rowSpan + 2, rowSpan + 2);

    for (let row = 0; row <= rowLimit; row++) {
      for (let col = 0; col <= columns - colSpan; col++) {
        if (!canPlace(usedRows, columns, col, row, colSpan, rowSpan)) continue;

        const candidate = { col, row, colSpan, rowSpan };
        const lastTouch = last ? touchLength(candidate, last) : 0;
        const allTouch = placed.reduce((sum, rect) => sum + touchLength(candidate, rect), 0);
        const adjacent = adjacencyScore(usedRows, candidate);
        const distance = last ? rectDistance(candidate, last) : 0;
        const alignment = alignmentScore(candidate, placed);
        const horizontal = horizontalScore(candidate, placed);
        const newBottom = Math.max(maxRow, row + rowSpan);
        const xReach = col + colSpan;
        const score = placed.length
          ? row * 100000 + col * 1000 + newBottom * 120 + distance * 60 - horizontal * 9000 - colSpan * 2600 - xReach * 220 - lastTouch * 800 - allTouch * 500 - adjacent * 300 - alignment * 80
          : row * 100 + col;

        if (!best || score < best.score) {
          best = { col, row, colSpan, rowSpan, score };
        }
      }
    }

    return best || { col: 0, row: maxRow, colSpan, rowSpan, score: 0 };
  }

  function setShapeClasses(card, colSpan, rowSpan) {
    card.classList.toggle('task-card--compact', colSpan <= 2);
    card.classList.toggle('task-card--wide', colSpan >= 4);
    card.classList.toggle('task-card--tall', rowSpan >= 3);
  }

  function getBaseShape(columns, finalCol) {
    const preferredColSpan = columns <= 4 ? 2 : 3;
    return {
      colSpan: Math.max(1, Math.min(preferredColSpan, columns - finalCol)),
      rowSpan: 2,
    };
  }

  async function render(options) {
    const config = Object.assign({}, DEFAULT_OPTIONS, options || {});
    const board = config.board;
    const items = config.items || [];
    const createCard = config.createCard;
    const isCancelled = config.isCancelled || function() { return false; };

    if (!board || typeof createCard !== 'function') return;
    if (isCancelled()) return;

    board.innerHTML = '';
    board.style.height = '0px';
    board.setAttribute('aria-busy', 'true');

    if (!items.length) {
      board.innerHTML = '<div class="tasks-empty">Задач нет</div>';
      board.style.height = '';
      board.removeAttribute('aria-busy');
      return;
    }

    const boardWidth = board.clientWidth || window.innerWidth;
    const columns = getColumns(boardWidth);
    const cellSize = Math.floor((boardWidth - config.gap * (columns - 1)) / columns);
    const usedRows = [];
    const placed = [];
    const settlePromises = [];
    let maxBottom = 0;
    let maxRow = 0;

    for (let index = 0; index < items.length; index++) {
      if (isCancelled()) return;

      const item = items[index];
      const card = createCard(item);
      board.appendChild(card);

      let best = null;
      const shapes = getShapeOptions(item, index, columns);

      for (let shapeIndex = 0; shapeIndex < shapes.length; shapeIndex++) {
        const [colSpan, baseRowSpan] = shapes[shapeIndex];
        const width = cellSize * colSpan + config.gap * (colSpan - 1);

        card.style.width = width + 'px';
        card.style.height = 'auto';
        card.style.visibility = 'hidden';
        setShapeClasses(card, colSpan, baseRowSpan);

        const measuredHeight = Math.max(card.offsetHeight, card.scrollHeight);
        const rowSpan = Math.max(baseRowSpan, Math.ceil((measuredHeight + config.gap) / (cellSize + config.gap)));
        const place = findBestPlace(usedRows, columns, colSpan, rowSpan, placed, maxRow);
        const score = place.score + shapeIndex * 0.2;

        if (!best || score < best.score || (score === best.score && colSpan > best.colSpan)) {
          best = { colSpan, width, rowSpan, place, score };
        }
      }

      const left = best.place.col * (cellSize + config.gap);
      const top = best.place.row * (cellSize + config.gap);
      const height = cellSize * best.rowSpan + config.gap * (best.rowSpan - 1);
      const baseShape = getBaseShape(columns, best.place.col);
      const baseWidth = cellSize * baseShape.colSpan + config.gap * (baseShape.colSpan - 1);
      const baseHeight = cellSize * baseShape.rowSpan + config.gap * (baseShape.rowSpan - 1);

      markPlace(usedRows, best.place.col, best.place.row, best.colSpan, best.rowSpan);
      placed.push({ col: best.place.col, row: best.place.row, colSpan: best.colSpan, rowSpan: best.rowSpan });

      card.style.left = left + 'px';
      card.style.top = top + 'px';
      card.style.width = baseWidth + 'px';
      card.style.height = baseHeight + 'px';
      card.style.visibility = '';
      card.classList.add('task-card--animating');
      setShapeClasses(card, baseShape.colSpan, baseShape.rowSpan);

      maxBottom = Math.max(maxBottom, top + height);
      maxRow = Math.max(maxRow, best.place.row + best.rowSpan);

      await nextFrame();
      if (isCancelled()) return;

      card.classList.add('task-card--visible');

      await nextFrame();
      if (isCancelled()) return;

      setShapeClasses(card, best.colSpan, baseShape.rowSpan);
      card.style.width = best.width + 'px';

      await wait(config.widthDuration);
      if (isCancelled()) return;

      setShapeClasses(card, best.colSpan, best.rowSpan);
      card.style.height = height + 'px';
      board.style.height = maxBottom + 'px';

      settlePromises.push(wait(config.heightDuration).then(function() {
        if (isCancelled()) return;
        card.classList.remove('task-card--animating');
        card.classList.add('task-card--settled');
      }));

      await wait(config.stepDelay);
    }

    await Promise.all(settlePromises);
    if (isCancelled()) return;

    board.removeAttribute('aria-busy');
  }

  window.TaskTileLayout = { render };
})();
