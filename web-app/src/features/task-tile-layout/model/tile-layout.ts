const DEFAULT_OPTIONS = {
  gap: 8,
};

export interface TileLayoutItem {
  title?: string;
  description?: string | null;
  tags?: unknown[];
  level?: string;
}

export interface TileLayoutOptions<TItem extends TileLayoutItem> {
  items: TItem[];
  boardWidth: number;
  gap?: number;
}

export interface TileLayoutTile<TItem> {
  item: TItem;
  index: number;
  left: number;
  top: number;
  width: number;
  height: number;
  baseWidth: number;
  baseHeight: number;
  colSpan: number;
  rowSpan: number;
  compact: boolean;
  wide: boolean;
  tall: boolean;
}

export interface TileLayoutResult<TItem> {
  tiles: TileLayoutTile<TItem>[];
  height: number;
  columns: number;
}

function getColumns(width: number): number {
  if (width < 560) return 4;
  if (width < 900) return 6;
  if (width < 1180) return 8;
  return 12;
}

function uniqueShapes(shapes: [number, number][], columns: number): [number, number][] {
  const seen: Record<string, boolean> = {};
  return shapes
    .filter(([w, h]) => {
      const width = Math.min(Math.max(w, 2), columns);
      const key = `${width}x${h}`;
      if (seen[key]) return false;
      seen[key] = true;
      return width > 0 && h > 0;
    })
    .map(([w, h]) => [Math.min(Math.max(w, 2), columns), h]);
}

function getContentWeight(item: TileLayoutItem): number {
  return (
    String(item.title || "").length +
    Math.round(String(item.description || "").length * 0.35) +
    (item.tags || []).length * 8
  );
}

function getShapeOptions(item: TileLayoutItem, index: number, columns: number): [number, number][] {
  const mobileShapes: [number, number][] = [
    [4, 2],
    [2, 2],
    [2, 3],
    [4, 2],
    [2, 2],
    [2, 3],
  ];
  const smallShapes: [number, number][] = [
    [3, 3],
    [3, 2],
    [3, 3],
    [2, 3],
    [4, 2],
    [3, 3],
  ];
  const tabletShapes: [number, number][] = [
    [4, 3],
    [4, 2],
    [4, 3],
    [3, 3],
    [4, 2],
    [3, 3],
  ];
  const desktopShapes: [number, number][] = [
    [4, 3],
    [3, 2],
    [5, 2],
    [2, 3],
    [4, 2],
    [3, 3],
  ];
  const shapes =
    columns <= 4
      ? mobileShapes
      : columns <= 6
        ? smallShapes
        : columns <= 8
          ? tabletShapes
          : desktopShapes;
  const primary = shapes[index % shapes.length];
  const contentWeight = getContentWeight(item);
  const needsRoom = item.level === "senior" || item.level === "Senior" || contentWeight > 86;

  return uniqueShapes(
    [
      primary,
      needsRoom ? [primary[0] + 1, primary[1]] : [primary[0], primary[1] + 1],
      [primary[0] - 1, primary[1] + 1],
    ],
    columns,
  );
}

interface Rect {
  col: number;
  row: number;
  colSpan: number;
  rowSpan: number;
}

function canPlace(
  usedRows: boolean[][],
  columns: number,
  col: number,
  row: number,
  colSpan: number,
  rowSpan: number,
): boolean {
  if (col + colSpan > columns) return false;
  for (let r = row; r < row + rowSpan; r++) {
    for (let c = col; c < col + colSpan; c++) {
      if (usedRows[r]?.[c]) return false;
    }
  }
  return true;
}

function markPlace(
  usedRows: boolean[][],
  col: number,
  row: number,
  colSpan: number,
  rowSpan: number,
): void {
  for (let r = row; r < row + rowSpan; r++) {
    if (!usedRows[r]) usedRows[r] = [];
    for (let c = col; c < col + colSpan; c++) {
      usedRows[r][c] = true;
    }
  }
}

function touchLength(a: Rect, b: Rect): number {
  let touch = 0;
  if (a.col + a.colSpan === b.col || b.col + b.colSpan === a.col) {
    touch += Math.max(0, Math.min(a.row + a.rowSpan, b.row + b.rowSpan) - Math.max(a.row, b.row));
  }
  if (a.row + a.rowSpan === b.row || b.row + b.rowSpan === a.row) {
    touch += Math.max(0, Math.min(a.col + a.colSpan, b.col + b.colSpan) - Math.max(a.col, b.col));
  }
  return touch;
}

function rectDistance(a: Rect, b: Rect): number {
  const dx =
    a.col + a.colSpan < b.col
      ? b.col - (a.col + a.colSpan)
      : b.col + b.colSpan < a.col
        ? a.col - (b.col + b.colSpan)
        : 0;
  const dy =
    a.row + a.rowSpan < b.row
      ? b.row - (a.row + a.rowSpan)
      : b.row + b.rowSpan < a.row
        ? a.row - (b.row + b.rowSpan)
        : 0;
  return dx + dy;
}

function adjacencyScore(usedRows: boolean[][], candidate: Rect): number {
  let score = 0;
  const left = candidate.col - 1;
  const right = candidate.col + candidate.colSpan;
  const top = candidate.row - 1;
  const bottom = candidate.row + candidate.rowSpan;
  for (let r = candidate.row; r < candidate.row + candidate.rowSpan; r++) {
    if (usedRows[r]?.[left]) score++;
    if (usedRows[r]?.[right]) score++;
  }
  for (let c = candidate.col; c < candidate.col + candidate.colSpan; c++) {
    if (usedRows[top]?.[c]) score++;
    if (usedRows[bottom]?.[c]) score++;
  }
  return score;
}

function alignmentScore(candidate: Rect, placed: Rect[]): number {
  let score = 0;
  for (const rect of placed) {
    if (candidate.col === rect.col) score++;
    if (candidate.row === rect.row) score++;
    if (candidate.col + candidate.colSpan === rect.col + rect.colSpan) score++;
    if (candidate.row + candidate.rowSpan === rect.row + rect.rowSpan) score++;
  }
  return score;
}

function horizontalScore(candidate: Rect, placed: Rect[]): number {
  let score = 0;
  const last = placed[placed.length - 1];
  if (last && candidate.row === last.row && candidate.col === last.col + last.colSpan) score += 12;
  if (last && candidate.row === last.row && candidate.col > last.col) score += 6;
  for (const rect of placed) {
    if (candidate.row === rect.row && candidate.col === rect.col + rect.colSpan) score += 5;
    if (candidate.row === rect.row && candidate.col + candidate.colSpan === rect.col) score += 3;
  }
  return score;
}

function findBestPlace(
  usedRows: boolean[][],
  columns: number,
  colSpan: number,
  rowSpan: number,
  placed: Rect[],
  maxRow: number,
): Rect & { score: number } {
  let best: (Rect & { score: number }) | null = null;
  const last = placed[placed.length - 1];
  const rowLimit = Math.max(maxRow + rowSpan + 2, rowSpan + 2);

  for (let row = 0; row <= rowLimit; row++) {
    for (let col = 0; col <= columns - colSpan; col++) {
      if (!canPlace(usedRows, columns, col, row, colSpan, rowSpan)) continue;

      const candidate: Rect = { col, row, colSpan, rowSpan };
      const lastTouch = last ? touchLength(candidate, last) : 0;
      const allTouch = placed.reduce((sum, rect) => sum + touchLength(candidate, rect), 0);
      const adjacent = adjacencyScore(usedRows, candidate);
      const distance = last ? rectDistance(candidate, last) : 0;
      const alignment = alignmentScore(candidate, placed);
      const horizontal = horizontalScore(candidate, placed);
      const newBottom = Math.max(maxRow, row + rowSpan);
      const xReach = col + colSpan;
      const score = placed.length
        ? row * 100000 +
          col * 1000 +
          newBottom * 120 +
          distance * 60 -
          horizontal * 9000 -
          colSpan * 2600 -
          xReach * 220 -
          lastTouch * 800 -
          allTouch * 500 -
          adjacent * 300 -
          alignment * 80
        : row * 100 + col;

      if (!best || score < best.score) {
        best = { col, row, colSpan, rowSpan, score };
      }
    }
  }

  return best || { col: 0, row: maxRow, colSpan, rowSpan, score: 0 };
}

function estimateRowSpan(item: TileLayoutItem, colSpan: number, baseRowSpan: number): number {
  const contentWeight = getContentWeight(item);
  const widthRelief = Math.max(0, colSpan - 2) * 16;
  const adjustedWeight = Math.max(0, contentWeight - widthRelief);
  const extraRows = Math.max(0, Math.ceil((adjustedWeight - 96) / 54));

  return Math.max(baseRowSpan, Math.min(baseRowSpan + extraRows, baseRowSpan + 2));
}

function getBaseShape(columns: number, finalCol: number) {
  const preferredColSpan = columns <= 4 ? 2 : 3;
  return {
    colSpan: Math.max(1, Math.min(preferredColSpan, columns - finalCol)),
    rowSpan: 2,
  };
}

export function calculateTileLayout<TItem extends TileLayoutItem>(
  options: TileLayoutOptions<TItem>,
): TileLayoutResult<TItem> {
  const config = { ...DEFAULT_OPTIONS, ...options };
  const { items, boardWidth } = config;

  if (!items.length || boardWidth <= 0) {
    return { tiles: [], height: 0, columns: getColumns(boardWidth) };
  }

  const columns = getColumns(boardWidth);
  const cellSize = Math.floor((boardWidth - config.gap * (columns - 1)) / columns);
  const usedRows: boolean[][] = [];
  const placed: Rect[] = [];
  const tiles: TileLayoutTile<TItem>[] = [];
  let maxBottom = 0;
  let maxRow = 0;

  for (let index = 0; index < items.length; index++) {
    const item = items[index];
    const shapes = getShapeOptions(item, index, columns);
    let best: {
      colSpan: number;
      rowSpan: number;
      place: Rect;
      score: number;
    } | null = null;

    for (let shapeIndex = 0; shapeIndex < shapes.length; shapeIndex++) {
      const [colSpan, baseRowSpan] = shapes[shapeIndex];
      const rowSpan = estimateRowSpan(item, colSpan, baseRowSpan);
      const place = findBestPlace(usedRows, columns, colSpan, rowSpan, placed, maxRow);
      const score = place.score + shapeIndex * 0.2;

      if (!best || score < best.score || (score === best.score && colSpan > best.colSpan)) {
        best = { colSpan, rowSpan, place, score };
      }
    }

    if (!best) continue;

    const left = best.place.col * (cellSize + config.gap);
    const top = best.place.row * (cellSize + config.gap);
    const width = cellSize * best.colSpan + config.gap * (best.colSpan - 1);
    const height = cellSize * best.rowSpan + config.gap * (best.rowSpan - 1);
    const baseShape = getBaseShape(columns, best.place.col);
    const baseWidth = cellSize * baseShape.colSpan + config.gap * (baseShape.colSpan - 1);
    const baseHeight = cellSize * baseShape.rowSpan + config.gap * (baseShape.rowSpan - 1);

    markPlace(usedRows, best.place.col, best.place.row, best.colSpan, best.rowSpan);
    placed.push({
      col: best.place.col,
      row: best.place.row,
      colSpan: best.colSpan,
      rowSpan: best.rowSpan,
    });

    maxBottom = Math.max(maxBottom, top + height);
    maxRow = Math.max(maxRow, best.place.row + best.rowSpan);

    tiles.push({
      item,
      index,
      left,
      top,
      width,
      height,
      baseWidth,
      baseHeight,
      colSpan: best.colSpan,
      rowSpan: best.rowSpan,
      compact: best.colSpan <= 2,
      wide: best.colSpan >= 4,
      tall: best.rowSpan >= 3,
    });
  }

  return { tiles, height: maxBottom, columns };
}
