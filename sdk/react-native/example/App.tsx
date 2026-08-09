import React, { useEffect } from 'react';
import { AppState, Button, SafeAreaView } from 'react-native';
import {
  MobileLabClient,
  MobileLabLifecycleReporter,
} from '@mobilelab/react-native';

const client = new MobileLabClient({ endpoint: 'http://10.0.2.2:4566' });

export default function App() {
  useEffect(() => {
    const reporter = new MobileLabLifecycleReporter(client, AppState);
    reporter.attach();
    return () => reporter.detach();
  }, []);

  return (
    <SafeAreaView>
      <Button
        title="Report MobileLab marker"
        onPress={() => void client.marker('example.button_pressed')}
      />
    </SafeAreaView>
  );
}
