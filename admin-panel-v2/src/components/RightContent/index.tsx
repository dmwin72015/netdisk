import { BookOutlined } from '@ant-design/icons';
import { Button, Tooltip } from 'antd';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { LangDropdown } from './LangDropdown';
import useHeaderActionStyles from './style';
import { VersionDropdown } from './VersionDropdown';

export const DocLink: React.FC = () => {
  const { styles } = useHeaderActionStyles();
  const navigate = useNavigate();

  return (
    <Tooltip title="使用文档">
      <Button
        type="text"
        className={styles.action}
        icon={<BookOutlined />}
        aria-label="使用文档"
        onClick={() => {
          navigate('/welcome');
        }}
      />
    </Tooltip>
  );
};

export { LangDropdown, VersionDropdown };
