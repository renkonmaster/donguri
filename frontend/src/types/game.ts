export type Player = {
  id: string;
  name: string;
  orderIndex: number;
  lat: number;
  lng: number;
};

export type Message = {
  id: string;
  senderId: string;
  receiverId: string;
  content: string;
  createdAt: Date;
};
