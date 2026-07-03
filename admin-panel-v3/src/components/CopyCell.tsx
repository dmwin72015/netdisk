import { Typography } from "antd";

const { Text } = Typography;

interface CopyCellProps {
  value: string;
  children?: React.ReactNode;
}

export default function CopyCell({ value, children }: CopyCellProps) {
  return (
    <Text
      copyable={{ text: value, tooltips: ["复制", "已复制"] }}
      className=" truncate"
    >
      {children ?? value}
    </Text>
  );
}
