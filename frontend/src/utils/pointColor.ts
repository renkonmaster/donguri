export type Rgb = [number, number, number];
export type Role = 'highlight' | 'adjacent' | 'normal';

// ID を HSL の RGB に変換する
// 黄金角 (137.508°) を掛けることで、連番 ID でも hue が均等に散らばる
// role に応じて彩度・明度を調整する: highlight は鮮やか、adjacent は通常、normal は淡め
export function idToRgb(id: string, role: Role): Rgb {
  const n = [...id].reduce((h, c) => (h * 31 + c.charCodeAt(0)) | 0, 0);
  const h = Math.abs(n) * 137.508 % 360;
  const [s, l] = role === 'highlight'
    ? [0.90, 0.48]
    : role === 'adjacent'
      ? [0.75, 0.55]
      : [0.45, 0.65];
  const a = s * Math.min(l, 1 - l);
  const channel = (pos: number) => {
    const k = (pos + h / 30) % 12;
    return Math.round((l - a * Math.max(-1, Math.min(k - 3, Math.min(9 - k, 1)))) * 255);
  };
  return [channel(0), channel(8), channel(4)];
}

export function lerpRgb(a: Rgb, b: Rgb, t: number): Rgb {
  return [
    Math.round(a[0] + (b[0] - a[0]) * t),
    Math.round(a[1] + (b[1] - a[1]) * t),
    Math.round(a[2] + (b[2] - a[2]) * t),
  ];
}
