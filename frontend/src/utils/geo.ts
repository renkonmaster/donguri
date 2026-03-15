// 2 点間の大圏コース (測地線) を steps 個の線分で補間した座標列を返す
export function greatCircleSegment(
  from: { lat: number; lng: number },
  to: { lat: number; lng: number },
  steps = 64,
): [number, number][] {
  steps = Math.max(1, steps);
  const toRad = (d: number) => (d * Math.PI) / 180;
  const toDeg = (r: number) => (r * 180) / Math.PI;
  const lat1 = toRad(from.lat), lng1 = toRad(from.lng);
  const lat2 = toRad(to.lat), lng2 = toRad(to.lng);
  const d = 2 * Math.asin(Math.sqrt(
    Math.sin((lat2 - lat1) / 2) ** 2
    + Math.cos(lat1) * Math.cos(lat2) * Math.sin((lng2 - lng1) / 2) ** 2,
  ));
  if (d === 0) return [[from.lng, from.lat]];
  return Array.from({ length: steps + 1 }, (_, i) => {
    const t = i / steps;
    const A = Math.sin((1 - t) * d) / Math.sin(d);
    const B = Math.sin(t * d) / Math.sin(d);
    const x = A * Math.cos(lat1) * Math.cos(lng1) + B * Math.cos(lat2) * Math.cos(lng2);
    const y = A * Math.cos(lat1) * Math.sin(lng1) + B * Math.cos(lat2) * Math.sin(lng2);
    const z = A * Math.sin(lat1) + B * Math.sin(lat2);
    return [toDeg(Math.atan2(y, x)), toDeg(Math.atan2(z, Math.sqrt(x * x + y * y)))];
  });
}

// 日付変更線越えで経度が飛ばないよう、前の点との差が±180°以内になるよう補正する
export function unwrapLongitudes(coords: [number, number][]): [number, number][] {
  const result: [number, number][] = [];
  for (const [rawLng, lat] of coords) {
    let lng = rawLng;
    if (result.length > 0) {
      const prevLng = result[result.length - 1][0];
      while (lng - prevLng > 180) lng -= 360;
      while (prevLng - lng > 180) lng += 360;
    }
    result.push([lng, lat]);
  }
  return result;
}
