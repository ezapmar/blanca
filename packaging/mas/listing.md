# Blanca on the Mac App Store

What goes into App Store Connect, field by field. `scripts/macapp --store` builds the
package that goes with it.

## App information

- **Name:** Blanca
- **Subtitle** (30 at most): Clipboard history, one hotkey
- **Category:** Productivity. Secondary: Utilities
- **Price:** Free
- **Privacy Policy URL:** https://ezapmar.github.io/blanca/privacy.html
- **Support URL:** https://github.com/ezapmar/blanca/issues
- **Marketing URL:** https://ezapmar.github.io/blanca/
- **Copyright:** 2026 Tunca Üçer

## Promotional text (170 at most)

Everything you copy, kept. One hotkey brings it back, one clipping at a time, and letting go of the keys pastes it where you were.

## Keywords (100 at most, commas, no spaces)

clipboard,history,manager,paste,copy,clip,pasteboard,snippet,hotkey,shortcut,menu bar,keyboard

## Description

Blanca remembers the text you copy, so the thing you copied ten minutes ago is still there when you need it.

Press Control+Option+V and a bezel appears in the middle of the screen with your latest clipping. Press V again to step back through older ones. Let go of the keys and Blanca pastes the one you stopped on, right where you were typing. Nothing to click, no window to manage.

ONE HOTKEY
• Control+Option+V steps to older clippings, with Shift to newer ones.
• Arrow keys, Home, End, Page Up and Page Down move through the list.
• 1 to 9 jump straight to that clipping, 0 to the tenth.
• Delete forgets the clipping on screen. Escape closes the bezel and changes nothing.
• Prefer to take your time? Switch on Sticky bezel and choose with Return.

A MENU IN THE BAR
• Click the dog in the menu bar for your newest clippings. Click one to use it.
• Clear the last hour, the last 24 hours, the last month, or everything.
• Every setting is a switch in the same menu. There is no separate settings window.

MADE TO STAY OUT OF THE WAY
• Lives in the menu bar. No Dock icon, no window.
• Remembers up to 200 clippings. You choose how many, and how many the menu shows.
• Text copied on your iPhone or iPad arrives through Universal Clipboard, marked with a small phone.
• Opens at login if you want it to.

PRIVATE BY DESIGN
• Your clippings stay on your Mac. Blanca has no account, no sync, no analytics, and never connects to the internet.
• Copies from password managers are skipped by default.
• Very large copies are skipped by default, so a stray 50,000 characters do not fill the list.

Blanca keeps text. Images and files you copy are left alone.

To paste for you, Blanca needs to be allowed under System Settings > Privacy & Security > Accessibility. It asks once, the first time you open it. Without that permission it still puts the clipping you choose on the clipboard, and you paste it yourself with Command+V.

Blanca is open source under the MIT licence, and named after a dog.

## What's New (first version)

Blanca's first release on the Mac App Store.

## App Privacy (the questionnaire)

- **Do you or your third-party partners collect data from this app?** No.
- That gives the label **Data Not Collected**. Nothing else to fill in.

## Age rating

Answer None or No to every question. The rating comes out as 4+.

## Export compliance

The app uses no encryption of its own; `ITSAppUsesNonExemptEncryption` is false in its
Info.plist, so App Store Connect does not ask.

## App Review information

- **Sign-in required:** No.
- **Contact:** Tunca Üçer, tuncaucer@gmail.com

### Notes for the reviewer

Blanca is a clipboard history manager that lives in the menu bar. It has no Dock icon and no main window: look for the dog's head in the menu bar after launch.

How to try it:
1. Open Blanca. macOS asks to allow it under Privacy & Security > Accessibility. Please allow it.
2. Copy a few pieces of text in any app, TextEdit for example.
3. Click into a text field and press Control+Option+V. A bezel appears with the latest clipping. Keep Control+Option held and press V again to step to older ones.
4. Let go of the keys. Blanca pastes the clipping that was on screen.
5. Click the dog in the menu bar to see the same clippings as a menu, with Clear, Settings and About.

Why Accessibility is needed: after you choose a clipping, Blanca places it on the clipboard and sends one Command+V keystroke to the frontmost app, so the text lands where you were typing. That keystroke is the only thing the permission is used for. Blanca does not read the screen, other apps' windows or their content, and does not record or log keystrokes. If the permission is declined, the app still works: the chosen clipping is put on the clipboard and the user pastes it with Command+V.

Clipboard access: Blanca checks the general pasteboard's change count twice a second and keeps plain text only. Items a password manager marks as concealed are skipped. The history is a file inside the app's sandbox container. The app makes no network connections and has no network entitlement.

The global hotkey is registered with RegisterEventHotKey, which needs no extra permission.
