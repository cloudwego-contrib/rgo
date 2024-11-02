import org.jetbrains.intellij.platform.gradle.extensions.intellijPlatform

plugins {
    id("org.jetbrains.intellij.platform") version "2.1.0"
    id("java")
}

repositories {
    mavenCentral()

    intellijPlatform {
        defaultRepositories()
    }

}

dependencies {
    intellijPlatform {
        intellijIdeaUltimate("2024.2.3")
        bundledPlugin("com.intellij.java")

        pluginVerifier()
        zipSigner()
        instrumentationTools()
    }
}