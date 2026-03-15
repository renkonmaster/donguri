export type MapPoint = {
  id: string;
  lat: number;
  lng: number;
  name?: string;
};

export type MapClickPayload = {
  lat: number;
  lng: number;
  point?: MapPoint;
};
