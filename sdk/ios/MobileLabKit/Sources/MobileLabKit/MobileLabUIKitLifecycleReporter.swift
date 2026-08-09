#if canImport(UIKit)
import Foundation
import UIKit

@MainActor
public final class MobileLabUIKitLifecycleReporter: NSObject {
    private enum Event {
        case ready
        case foreground
        case background
    }

    private let reporter: MobileLabLifecycleReporter
    private let notificationCenter: NotificationCenter
    private let onError: @MainActor (Error) -> Void
    private var attached = false

    public init(
        client: MobileLabClient,
        notificationCenter: NotificationCenter = .default,
        onError: @escaping @MainActor (Error) -> Void = { _ in }
    ) {
        reporter = MobileLabLifecycleReporter(client: client)
        self.notificationCenter = notificationCenter
        self.onError = onError
        super.init()
    }

    public func attach(reportReady: Bool = true) {
        guard !attached else { return }
        attached = true
        notificationCenter.addObserver(
            self,
            selector: #selector(didBecomeActive),
            name: UIApplication.didBecomeActiveNotification,
            object: nil
        )
        notificationCenter.addObserver(
            self,
            selector: #selector(didEnterBackground),
            name: UIApplication.didEnterBackgroundNotification,
            object: nil
        )
        if reportReady { report(.ready) }
    }

    public func detach() {
        guard attached else { return }
        notificationCenter.removeObserver(self)
        attached = false
    }

    @objc private func didBecomeActive() {
        report(.foreground)
    }

    @objc private func didEnterBackground() {
        report(.background)
    }

    private func report(_ event: Event) {
        Task { @MainActor [weak self] in
            guard let self else { return }
            do {
                switch event {
                case .ready: try await reporter.reportReady()
                case .foreground: try await reporter.onForeground()
                case .background: try await reporter.onBackground()
                }
            } catch {
                onError(error)
            }
        }
    }
}
#endif
