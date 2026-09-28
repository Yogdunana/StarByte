import MemberImportModal from '@/components/member/MemberImportModal';
import PageIntro from '@/components/PageIntro/PageIntro';
import { tx, useLocale } from '@/i18n/text';
import React, { useState } from 'react';
import { Card, List, Space, Button } from 'antd';
import { useNavigate } from 'react-router-dom';

/**
 * 批量录入在系统管理里的入口，和「会员档案 → 批量录入」共用同一个弹窗。
 *
 * 放两处是有意的：档案页那个按钮是导完就地核对，这里的入口是给习惯先进管理后台的人。
 * 两边行为完全一致，所以录入逻辑只写一份。
 *
 * 下面每条规则的中文都直接写在 tx() 里，而且必须在组件内部调用：项目有一条单测
 * 兜着，凡是没被 tx() 包住的中文字面量会被判成硬编码；放到模块顶层求值的话，
 * 语种切换之后那些译文就过期了。
 */
const MemberImportPage: React.FC = () => {
  // useLocale 会让组件跟着语种切换重渲染，这里不需要它的返回值。
  useLocale();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);

  const notes = [
      {
        title: tx('学号是唯一凭据'),
        desc: tx('统一身份认证查不到绑定关系时会按学号回查档案，命中就自动绑到这个账号并登录。名单里学号填对，之后把登录地址发给他们就行，不用邀请码也不用先注册。'),
      },
      {
        title: tx('可录的角色'),
        desc: tx('会员、预备干事、干事、部长、副中心主任。会长、指导老师、中心主任不在白名单里 —— 要么唯一要么另有程序，不适合批量给。'),
      },
      {
        title: tx('部长和副中心主任必须带范围'),
        desc: tx('部长给部门（职能部门），副中心主任给中心。范围会一起写进 user_role_departments，否则带范围的审批环节选不到这个人。'),
      },
      {
        title: tx('协会职务要系统管理员'),
        desc: tx('录部长、副中心主任这类协会职务跟「角色管理 → 角色成员」同一条规则，只有系统管理员能做；其余角色有会员管理权限就够。逐行判定，一行没权限不影响同批其他人。'),
      },
      {
        title: tx('没写角色的按预备干事处理'),
        desc: tx('批量录入的人没有入会申请记录，预备期落在档案的 probation_until 上，定时任务每分钟扫一遍，满一个月自动转正式干事。'),
      },
      {
        title: tx('重跑同一份名单是幂等的'),
        desc: tx('按学号定位：内容没变就记为「无变化」，不做二次写入；已经停用或已经转正式的成员不会因为重复导入被降级。单次最多 500 行。'),
      },
    ];

  return (
    <div>
      <PageIntro
        eyebrow="SYSTEM / MEMBERS"
        title={tx('成员批量录入')}
        description={tx('测评系统出结果后一次性把名单录进来。这里的入口和「会员档案 → 批量录入」是同一个功能。')}
        actions={
          <Space>
            <Button onClick={() => navigate('/member/list')}>{tx('去会员档案核对')}</Button>
            <Button type="primary" onClick={() => setOpen(true)}>
              {tx('开始录入')}
            </Button>
          </Space>
        }
      />
      <Card>
        <List
          dataSource={notes}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta title={item.title} description={item.desc} />
            </List.Item>
          )}
        />
      </Card>
      <MemberImportModal open={open} onClose={() => setOpen(false)} onImported={() => undefined} />
    </div>
  );
};

export default MemberImportPage;
