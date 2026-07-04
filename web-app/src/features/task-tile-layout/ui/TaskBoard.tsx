import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  calculateTileLayout,
  type TileLayoutItem,
  type TileLayoutTile,
} from "../model/tile-layout";

const TILE_STEP_DELAY = 18;
const TILE_WIDTH_DURATION = 55;
const TILE_HEIGHT_DURATION = 65;

interface TaskBoardProps<TItem extends TileLayoutItem> {
  items: TItem[];
  getKey: (item: TItem) => string;
  renderItem: (item: TItem, tile: TileLayoutTile<TItem>) => ReactNode;
}

export function TaskBoard<TItem extends TileLayoutItem>({
  items,
  getKey,
  renderItem,
}: TaskBoardProps<TItem>) {
  const boardRef = useRef<HTMLDivElement>(null);
  const [boardWidth, setBoardWidth] = useState(0);
  const [animation, setAnimation] = useState({
    mounted: 0,
    visible: 0,
    width: 0,
    height: 0,
    settled: 0,
  });

  useEffect(() => {
    const board = boardRef.current;
    if (!board) return;

    const setStableBoardWidth = (width: number) => {
      const nextWidth = Math.floor(width);
      setBoardWidth((prevWidth) => (Math.abs(prevWidth - nextWidth) <= 1 ? prevWidth : nextWidth));
    };

    const updateWidth = () => setStableBoardWidth(board.clientWidth);
    updateWidth();

    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry) return;
      setStableBoardWidth(entry.contentRect.width);
    });
    observer.observe(board);

    return () => observer.disconnect();
  }, []);

  const layout = useMemo(() => calculateTileLayout({ items, boardWidth }), [boardWidth, items]);

  useEffect(() => {
    let cancelled = false;
    const timers: number[] = [];
    const frames: number[] = [];
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    const wait = (ms: number) =>
      new Promise<void>((resolve) => {
        const timer = window.setTimeout(resolve, ms);
        timers.push(timer);
      });
    const nextFrame = () =>
      new Promise<void>((resolve) => {
        const frame = requestAnimationFrame(() => resolve());
        frames.push(frame);
      });

    if (!layout.tiles.length || reduceMotion) {
      const all = layout.tiles.length;
      setAnimation({ mounted: all, visible: all, width: all, height: all, settled: all });
      return () => {};
    }

    setAnimation({ mounted: 0, visible: 0, width: 0, height: 0, settled: 0 });

    (async () => {
      for (let index = 0; index < layout.tiles.length; index++) {
        if (cancelled) return;
        const count = index + 1;

        setAnimation((state) => ({ ...state, mounted: count }));
        await nextFrame();
        if (cancelled) return;

        setAnimation((state) => ({ ...state, visible: count }));
        await nextFrame();
        if (cancelled) return;

        setAnimation((state) => ({ ...state, width: count }));
        await wait(TILE_WIDTH_DURATION);
        if (cancelled) return;

        setAnimation((state) => ({ ...state, height: count }));
        await wait(TILE_HEIGHT_DURATION);
        if (cancelled) return;

        setAnimation((state) => ({ ...state, settled: count }));
        await wait(TILE_STEP_DELAY);
      }
    })();

    return () => {
      cancelled = true;
      for (const frame of frames) cancelAnimationFrame(frame);
      for (const timer of timers) window.clearTimeout(timer);
    };
  }, [layout]);

  if (!items.length) {
    return (
      <div ref={boardRef} className="tasks-board" aria-live="polite">
        <div className="tasks-empty">Задач нет</div>
      </div>
    );
  }

  return (
    <div
      ref={boardRef}
      className="tasks-board"
      style={{ height: layout.height || undefined }}
      aria-live="polite"
    >
      {layout.tiles.slice(0, animation.mounted).map((tile) => (
        <div
          key={getKey(tile.item)}
          className={`tasks-board__tile ${tile.index < animation.visible ? "tasks-board__tile--visible" : ""} ${tile.index < animation.settled ? "tasks-board__tile--settled" : ""}`}
          style={{
            left: tile.left,
            top: tile.top,
            width: tile.index < animation.width ? tile.width : tile.baseWidth,
            height: tile.index < animation.height ? tile.height : tile.baseHeight,
          }}
        >
          {renderItem(tile.item, tile)}
        </div>
      ))}
    </div>
  );
}
