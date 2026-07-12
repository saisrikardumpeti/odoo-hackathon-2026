import axios from 'axios';

export interface Model {
  id: string;
  name: string;
}

export const fetchModels = async (): Promise<Model[]> => {
  const { data } = await axios.get('/api/models');
  return data;
};
