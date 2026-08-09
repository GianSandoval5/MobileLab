// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "MobileLabKit",
    platforms: [
        .iOS(.v15),
        .macOS(.v12),
    ],
    products: [
        .library(name: "MobileLabKit", targets: ["MobileLabKit"]),
    ],
    targets: [
        .target(name: "MobileLabKit"),
        .testTarget(name: "MobileLabKitTests", dependencies: ["MobileLabKit"]),
    ]
)
