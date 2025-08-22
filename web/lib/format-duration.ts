export const formatDuration = (seconds?: number) => {
  if (typeof seconds !== "number" || isNaN(seconds) || seconds < 0)
    return "--:--:--";
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  const pad = (n: number) => n.toString().padStart(2, "0");
  return `${pad(h)}:${pad(m)}:${pad(s)}`;
};
