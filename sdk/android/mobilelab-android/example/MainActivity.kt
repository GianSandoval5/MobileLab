package dev.mobilelab.example

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ProcessLifecycleOwner
import dev.mobilelab.MobileLabClient
import dev.mobilelab.MobileLabLifecycleReporter

class MainActivity : ComponentActivity() {
    private val mobileLab = MobileLabClient("http://10.0.2.2:4566")
    private val reporter = MobileLabLifecycleReporter(mobileLab)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        reporter.reportReady()
        ProcessLifecycleOwner.get().lifecycle.addObserver(object : DefaultLifecycleObserver {
            override fun onStart(owner: LifecycleOwner) = reporter.onForeground()
            override fun onStop(owner: LifecycleOwner) = reporter.onBackground()
        })
    }

    fun onExampleClicked() {
        Thread { mobileLab.marker("example.clicked") }.start()
    }
}
