export type MapPoint = {
  id: string;
  orderIndex: number;
  lat: number;
  lng: number;
  name?: string;
};

export type MapClickPayload = {
  lat: number;
  lng: number;
  point?: MapPoint;
};
