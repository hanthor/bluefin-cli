# Interactive Menu Structure

The Bluefin CLI provides a rich interactive menu system to manage your environment. Below is a diagram of the menu hierarchy and available options.

```mermaid
graph TD
    Main[Main Menu] --> Status[📊 Status]
    Main --> Bling[✨ Bling]
    Main --> MOTD[📰 MOTD]
    Main --> Bundles[📦 Install Tools]
    Main --> Wallpapers[🖼  Wallpapers]
    Main --> Starship[🚀 Starship Theme]

    Bling --> BlingAction{Action}
    BlingAction -->|Toggle Current| BlingToggle[Enable/Disable Current Shell]
    BlingAction -->|Configure Components| BlingComps[Select Tools]
    BlingAction -->|Manage Shells| BlingShells[Select Shells to Enable]

    BlingComps --> |Multi-Select| BlingToolsList[eza, ugrep, bat, atuin, starship, zoxide, uutils]
    BlingShells --> |Multi-Select| ShellsList[bash, zsh, fish]

    MOTD --> MOTDAction{Action}
    MOTDAction -->|Show| MOTDShow[Display MOTD]
    MOTDAction -->|Toggle| MOTDToggle[Select Shells to Enable]

    MOTDToggle --> |Multi-Select| ShellsList

    Bundles --> BundlesList[Select Bundles]
    BundlesList --> |Multi-Select| BundlesOptions[AI Tools, Artwork, CLI Essentials, CNCF Tools, Experimental IDE, Fonts, IDE Tools, K8s Tools]

    Wallpapers --> WallpapersList[Select Wallpapers]
    WallpapersList --> |Multi-Select| WallpaperCasks[List from ublue-os/tap]

    Starship --> StarshipThemes[Select Theme]
    StarshipThemes --> |Select| ThemeOptions[Nerd Font Symbols, Tokyo Night, Catppuccin Powerline, etc.]
```

## Section Descriptions

- **Status**: Checks the current configuration and installation status of tools.
- **Bling**: Manages shell enhancements like `eza`, `bat`, `starship`, etc. You can toggle them for specific shells or configure which tools are enabled.
- **MOTD**: Controls the "Message of the Day" that appears when you open a terminal.
- **Install Tools**: Allows you to install curated bundles of Homebrew packages for various use cases (AI, Dev, Kubernetes, etc.).
- **Wallpapers**: Browse and install wallpapers available as Homebrew casks.
- **Starship Theme**: Quickly switch between different presets for the Starship prompt.
