import { App as CapacitorApp } from '@capacitor/app';
import { useEffect } from 'react';

import { MobileLabClient, MobileLabLifecycleReporter } from '@mobilelab/capacitor';

const mobileLab = new MobileLabClient({ endpoint: 'http://10.0.2.2:4566' });
const lifecycle = new MobileLabLifecycleReporter(mobileLab, CapacitorApp);

export function App() {
  useEffect(() => {
    void lifecycle.attach();
    return () => { void lifecycle.detach(); };
  }, []);

  return <button onClick={() => void mobileLab.marker('example.clicked')}>Report marker</button>;
}
