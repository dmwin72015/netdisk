import { Spin } from 'antd';

export default function PageLoading() {
  return (
    <div className="flex justify-center items-center min-h-[50vh]">
      <Spin size="large" />
    </div>
  );
}
