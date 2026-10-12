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

Everything below goes in the Notes field, and the same text is the reply in Resolution
Center when review asks for it (guideline 2.1, Information Needed, as they did on the
first submission on 2026-10-10).

**What Blanca is, and who it is for.** Blanca is a clipboard history manager for macOS.
macOS keeps only the last thing you copied, so anything copied before that is gone: the
address you copied before the order number, the line of code you copied before the error
message. Blanca keeps the text you copy and brings it back with one keystroke, and pastes
the one you choose where you were typing. It is for anyone who types on a Mac all day and
copies more than one thing at a time: developers, writers, translators, support and data
entry work. It is free, has no accounts and no purchases.

**No Dock icon, no window.** Blanca is a menu bar app (LSUIElement). After launch, look
for the dog's head in the menu bar at the top right. There is no main window and no
settings window: the menu is the whole interface.

**Setting it up and seeing the main features.** No login, no demo account, no sample
files; nothing is locked or paid.

1. Open Blanca. macOS asks to allow it under System Settings > Privacy & Security >
   Accessibility. Please allow it. (If it is declined, the app still works; see below.)
2. Copy a few separate pieces of text in any app, TextEdit for example.
3. Click into a text field and hold Control+Option and press V. A bezel appears in the
   middle of the screen with the latest clipping. Keep Control+Option held and press V
   again to step back to older ones.
4. Let go of the keys. Blanca pastes the clipping that was on screen, at the cursor.
5. Click the dog in the menu bar: the same clippings as a menu (click one to use it),
   Clear (last hour, 24 hours, month, everything), the settings switches, and About.

**Why Accessibility is needed.** After a clipping is chosen, Blanca puts it on the
clipboard and sends one Command+V keystroke to the frontmost app, so the text lands where
you were typing. That keystroke is the only use of the permission. Blanca does not read
the screen, other apps' windows or their content, and does not record or log keystrokes.
With the permission declined, the chosen clipping is still placed on the clipboard and the
user pastes it with Command+V.

**External services, tools and platforms: none.** Blanca delivers all of its
functionality on the device, with Apple frameworks only (AppKit, Carbon's
RegisterEventHotKey for the global hotkey, NSPasteboard). There are no data providers, no
authentication services, no payment processors, no AI or machine learning services, no
analytics or crash reporting, and no third-party SDKs. The app makes no network
connections at all and ships with no network entitlement: it is sandboxed with
com.apple.security.app-sandbox alone. Clipboard history is a file inside the app's own
sandbox container and never leaves the device. Text copied on an iPhone or iPad can appear
in the list, but that is Apple's own Universal Clipboard delivering it to the Mac's
pasteboard; Blanca neither talks to the device nor to iCloud itself. The app is open
source under the MIT licence: https://github.com/ezapmar/blanca

**Regional differences: none.** The app behaves identically in every region and on every
network. There is no geofencing, no region-dependent content, no server to vary by
country. The interface is English only in this version.

**Regulated industries and third-party material: none.** Blanca is a utility; it is not
in a regulated industry and needs no licence or authorisation. It contains no third-party
content, trademarks, media or data. Every asset is original work by the developer: the
icon and the dog drawings are his own, of his own dog. The code is his, MIT licensed,
with no third-party dependencies.

**No account-based features, no user-generated content, no purchases.** There is no
registration, login or account, so there is nothing to delete and no credentials to
supply. Nothing is shared or published: a clipping is text the user copied on their own
Mac, visible only to them, never transmitted, so there is no content feed and nothing to
report or block. There are no in-app purchases or paid features; the whole app is free.
The App Privacy answer is Data Not Collected, and that is literally true: nothing is
collected or sent anywhere.

**Privacy protections built in.** Clippings a password manager marks as concealed are
skipped by default, and very large copies are skipped by default. Blanca keeps text only;
images and files that are copied are ignored. The pasteboard is polled twice a second for
its change count.

### The screen recording review asks for

Record on a Mac running the latest macOS (that is the physical device for a Mac app).
QuickTime Player > File > New Screen Recording, or Command+Shift+5. Keep it under five
minutes, no cuts, and start with the launch.

1. Start with Blanca not running. Open it from Applications and let the Accessibility
   prompt appear; open System Settings and switch Blanca on, on camera.
2. Point at the dog in the menu bar: say there is no Dock icon and no window.
3. In TextEdit, copy three separate lines, one after another.
4. Click into a text field, hold Control+Option, press V three times so the bezel steps
   back through all three, then let go and show the text being pasted at the cursor.
5. Click the dog: show the clippings in the menu, click one, then show Clear and the
   settings switches, and About.
6. Say aloud, or caption, that there is no account, no purchase, no user-generated
   content and no network connection, so none of those flows exist to show.

No resubmission is needed for this: reply in Resolution Center with the notes above and
attach the recording, and add the notes to the Notes field for future submissions.
