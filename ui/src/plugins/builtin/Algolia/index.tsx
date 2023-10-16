import { memo, FC } from 'react';
import { useTranslation } from 'react-i18next';

import pluginKit, { PluginInfo } from '@/utils/pluginKit';
import { SvgIcon } from '@/components';

import info from './info.yaml';
import { useGetAlgoliaInfo } from './services';
import './i18n';

const pluginInfo: PluginInfo = {
  slug_name: info.slug_name,
};

const Index: FC = () => {
  const { t } = useTranslation(pluginKit.getTransNs(), {
    keyPrefix: pluginKit.getTransKeyPrefix(pluginInfo),
  });

  const { data } = useGetAlgoliaInfo();
  console.log(data);
  if (!data?.icon) return null;

  return (
    <a
      className="d-flex align-items-center"
      href="https://www.algolia.com/"
      target="_blank"
      rel="noopener noreferrer">
      <span className="small text-secondary me-2">{t('search_by')}</span>
      <SvgIcon base64={data?.icon} />
    </a>
  );
};

export default {
  info: pluginInfo,
  component: memo(Index),
};
