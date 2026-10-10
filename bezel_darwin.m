// The macOS half of Blanca: pasteboard poller, global hotkey and bezel panel.
// Everything here runs on the main thread.
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#import <ServiceManagement/ServiceManagement.h>
#include "_cgo_export.h"

static const CGFloat side = 325; // the bezel is a square
static NSTextField *title, *body;

static BOOL hasType(NSArray<NSString *> *names) {
	for (NSString *t in NSPasteboard.generalPasteboard.types)
		if ([names containsObject:t]) return YES;
	return NO;
}

bool bzSensitive(void) {
	return hasType(@[@"PasswordPboardType", @"org.nspasteboard.ConcealedType", @"com.agilebits.onepassword"]);
}

void bzCopy(const char *text) {
	[NSPasteboard.generalPasteboard clearContents];
	[NSPasteboard.generalPasteboard setString:@(text) ?: @"" forType:NSPasteboardTypeString];
}

void bzClear(void) { [NSPasteboard.generalPasteboard clearContents]; }

// bzPaste sends Cmd+V to the frontmost app. macOS drops the events unless Accessibility is granted.
void bzPaste(void) {
	CGEventSourceRef src = CGEventSourceCreate(kCGEventSourceStateCombinedSessionState);
	for (int down = 1; down >= 0; down--) {
		CGEventRef e = CGEventCreateKeyboardEvent(src, kVK_ANSI_V, down);
		CGEventSetFlags(e, kCGEventFlagMaskCommand);
		CGEventPost(kCGAnnotatedSessionEventTap, e);
		CFRelease(e);
	}
	CFRelease(src);
}

// bzLogin reports whether Blanca.app is one of the user's Open at Login items, and
// bzSetLogin adds or removes it. Both are for the app bundle only.
bool bzLogin(void) { return SMAppService.mainAppService.status == SMAppServiceStatusEnabled; }

bool bzSetLogin(bool on) {
	NSError *err;
	SMAppService *s = SMAppService.mainAppService;
	BOOL ok = on ? [s registerAndReturnError:&err] : [s unregisterAndReturnError:&err];
	if (!ok) NSLog(@"blanca: open at login: %@", err.localizedDescription);
	return ok;
}

// bzAskMove asks, of a Blanca opened inside its disk image, whether to move it to dir.
bool bzAskMove(const char *dir) {
	[NSApplication sharedApplication];
	NSApp.activationPolicy = NSApplicationActivationPolicyAccessory;
	NSAlert *a = [NSAlert new];
	a.messageText = @"Move Blanca to the Applications folder?";
	a.informativeText = [NSString stringWithFormat:@"Blanca was opened from its disk image, and stops "
		@"working once that is ejected. Moving it puts Blanca.app in %s, in place of an older one, "
		@"and opens it from there.", dir];
	[a addButtonWithTitle:@"Move to Applications"];
	[a addButtonWithTitle:@"Quit"];
	[NSApp activateIgnoringOtherApps:YES];
	return [a runModal] == NSAlertFirstButtonReturn;
}

// Bezel takes the keyboard without activating Blanca, so closing it returns focus
// to the window the clipping is pasted into.
@interface Bezel : NSPanel
@end

@implementation Bezel
- (BOOL)canBecomeKeyWindow { return YES; }
- (void)keyDown:(NSEvent *)e {
	NSString *c = e.charactersIgnoringModifiers;
	goKey(e.keyCode, c.length ? [c characterAtIndex:0] : 0);
}
- (void)cancelOperation:(id)sender { goKey(kVK_Escape, 0); } // Ctrl+Esc never reaches keyDown
- (void)flagsChanged:(NSEvent *)e {
	NSEventModifierFlags held = NSEventModifierFlagControl | NSEventModifierFlagOption |
		NSEventModifierFlagCommand | NSEventModifierFlagShift;
	if (!(e.modifierFlags & held)) goRelease();
}
- (void)resignKeyWindow {
	[super resignKeyWindow];
	goKey(kVK_Escape, 0);
}
@end

static Bezel *bezel;

static NSTextField *label(NSRect frame, NSFont *font) {
	NSTextField *l = [NSTextField wrappingLabelWithString:@""];
	l.frame = frame;
	l.font = font;
	l.selectable = NO;
	l.cell.truncatesLastVisibleLine = YES;
	[bezel.contentView addSubview:l];
	return l;
}

static void build(void) {
	bezel = [[Bezel alloc] initWithContentRect:NSMakeRect(0, 0, side, side)
		styleMask:NSWindowStyleMaskBorderless | NSWindowStyleMaskNonactivatingPanel
		backing:NSBackingStoreBuffered defer:YES];
	bezel.level = NSModalPanelWindowLevel;
	bezel.opaque = NO;
	bezel.backgroundColor = NSColor.clearColor;
	bezel.appearance = [NSAppearance appearanceNamed:NSAppearanceNameVibrantDark];
	bezel.collectionBehavior = NSWindowCollectionBehaviorFullScreenAuxiliary |
		NSWindowCollectionBehaviorMoveToActiveSpace | NSWindowCollectionBehaviorIgnoresCycle;
	NSVisualEffectView *back = [[NSVisualEffectView alloc] initWithFrame:NSMakeRect(0, 0, side, side)];
	back.material = NSVisualEffectMaterialHUDWindow;
	back.state = NSVisualEffectStateActive;
	back.wantsLayer = YES;
	back.layer.cornerRadius = 25;
	bezel.contentView = back;
	title = label(NSMakeRect(20, side - 44, side - 40, 24), [NSFont boldSystemFontOfSize:14]);
	title.alignment = NSTextAlignmentCenter;
	body = label(NSMakeRect(20, 20, side - 40, side - 76), [NSFont systemFontOfSize:14]);
}

void bzShow(const char *text, const char *head) {
	title.stringValue = @(head) ?: @"";
	body.stringValue = @(text) ?: @"";
	if (bezel.visible) return;
	NSRect f = NSScreen.mainScreen.visibleFrame; // the screen with keyboard focus
	[bezel setFrameOrigin:NSMakePoint(NSMidX(f) - side / 2, NSMidY(f) - side / 2)];
	[bezel makeKeyAndOrderFront:nil];
}

void bzHide(void) { [bezel orderOut:nil]; }

// Menu is the menu bar item's menu: the newest clippings,
// Clear, Settings, About, Check for Updates and Quit. It is rebuilt from the history every time it opens.
@interface Menu : NSObject <NSMenuDelegate>
@end

static NSStatusItem *item;
static Menu *menu;
static NSMenu *prefs;  // the Settings submenu
static NSMenu *spans;  // the Clear submenu

@implementation Menu
- (void)menuNeedsUpdate:(NSMenu *)m {
	[m removeAllItems];
	goMenu();
	if (!m.numberOfItems) [m addItemWithTitle:@"<None>" action:nil keyEquivalent:@""];
	[m addItem:NSMenuItem.separatorItem];
	[m addItemWithTitle:@"Clear" action:nil keyEquivalent:@""].submenu = spans = [NSMenu new];
	goClearMenu();
	[m addItemWithTitle:@"Settings" action:nil keyEquivalent:@""].submenu = prefs = [NSMenu new];
	goSettings();
	[m addItemWithTitle:@"About Blanca" action:@selector(about:) keyEquivalent:@""].target = self;
#ifndef MAS // the App Store updates its Blanca
	[m addItemWithTitle:@"Check for Updates…" action:@selector(update:) keyEquivalent:@""].target = self;
#endif
	[m addItem:NSMenuItem.separatorItem];
	[m addItemWithTitle:@"Quit Blanca" action:@selector(terminate:) keyEquivalent:@""];
}
- (void)pick:(NSMenuItem *)i { goMenuPick((int)i.tag); }
- (void)set:(NSMenuItem *)i { goSet((int)i.tag); }
- (void)clear:(NSMenuItem *)i { // ask first
	NSAlert *a = [NSAlert new];
	a.messageText = [NSString stringWithFormat:@"Clear clippings: %@?", i.title.lowercaseString];
	[a addButtonWithTitle:@"Clear"];
	[a addButtonWithTitle:@"Cancel"];
	[NSApp activateIgnoringOtherApps:YES];
	if ([a runModal] == NSAlertFirstButtonReturn) goMenuClear((int)i.tag);
}
- (void)about:(id)sender { goAbout(); }
#ifndef MAS
- (void)update:(id)sender { // the check waits on the network, so not on the main thread
	dispatch_async(dispatch_get_global_queue(QOS_CLASS_USER_INITIATED, 0), ^{
		char *note = NULL, *tag = goUpdateCheck(&note);
		dispatch_async(dispatch_get_main_queue(), ^{
			NSAlert *a = [NSAlert new];
			if (tag) {
				a.messageText = [NSString stringWithFormat:@"Blanca %s is available", tag + 1];
				a.informativeText = @"Blanca restarts when it is installed. macOS then asks you to allow it under Accessibility again.";
				[a addButtonWithTitle:@"Update"];
				[a addButtonWithTitle:@"Later"];
			} else {
				a.messageText = @(note) ?: @"";
			}
			[NSApp activateIgnoringOtherApps:YES];
			if ([a runModal] == NSAlertFirstButtonReturn && tag) goUpdate(tag);
			free(tag);
			free(note);
		});
	});
}
#endif
@end

// bzAbout shows the name and version over a few words about Blanca and who made it, and
// offers its page and a letter to its maker.
void bzAbout(const char *head, const char *text, const char *url, const char *mail) {
	NSAlert *a = [NSAlert new];
	a.messageText = @(head) ?: @"";
	a.informativeText = @(text) ?: @"";
	[a addButtonWithTitle:@"OK"];
	[a addButtonWithTitle:@"Website"];
	[a addButtonWithTitle:@"Email"];
	NSURL *page = [NSURL URLWithString:@(url) ?: @""], *letter = [NSURL URLWithString:@(mail) ?: @""];
	[NSApp activateIgnoringOtherApps:YES];
	NSModalResponse r = [a runModal];
	if (r == NSAlertSecondButtonReturn && page) [NSWorkspace.sharedWorkspace openURL:page];
	if (r == NSAlertThirdButtonReturn && letter) [NSWorkspace.sharedWorkspace openURL:letter];
}

// bzUpdating writes "Updating…" beside the dog while the installer runs. Given why the
// update failed it takes that away again and says so. From any thread.
void bzUpdating(const char *failed) {
	NSString *why = failed ? @(failed) ?: @"" : nil;
	dispatch_async(dispatch_get_main_queue(), ^{
		item.length = why ? NSSquareStatusItemLength : NSVariableStatusItemLength;
		item.button.imagePosition = why ? NSImageOnly : NSImageLeft;
		item.button.title = why ? @"" : @" Updating…";
		if (!why) return;
		NSAlert *a = [NSAlert new];
		a.messageText = @"Blanca was not updated";
		a.informativeText = why;
		[NSApp activateIgnoringOtherApps:YES];
		[a runModal];
	});
}

// bzMenuAdd lists a clipping; one that came from another device gets a phone beside it.
void bzMenuAdd(const char *text, bool remote) {
	NSMenuItem *i = [item.menu addItemWithTitle:@(text) ?: @"" action:@selector(pick:) keyEquivalent:@""];
	if (remote) {
		i.image = [NSImage imageWithSystemSymbolName:@"iphone" accessibilityDescription:@"From another device"];
		// macOS 27 hides menu item images unless asked: NSMenuItemImageVisibilityVisible.
		// Set by key so this still builds against an older SDK.
		if ([i respondsToSelector:NSSelectorFromString(@"setPreferredImageVisibility:")])
			[i setValue:@1 forKey:@"preferredImageVisibility"];
	}
	i.target = menu;
	i.tag = item.menu.numberOfItems - 1;
}

// bzClearAdd adds a line to the Clear submenu: how far back to clear.
void bzClearAdd(const char *text) {
	NSMenuItem *i = [spans addItemWithTitle:@(text) ?: @"" action:@selector(clear:) keyEquivalent:@""];
	i.target = menu;
	i.tag = spans.numberOfItems - 1;
}

// bzSetting adds a line to the Settings submenu or, with sub, to the last line's own submenu.
void bzSetting(const char *text, bool on, int tag, bool sub) {
	NSMenu *to = prefs;
	if (sub) {
		NSMenuItem *last = prefs.itemArray.lastObject;
		to = last.submenu ?: (last.submenu = [NSMenu new]);
	}
	NSMenuItem *i = [to addItemWithTitle:@(text) ?: @"" action:@selector(set:) keyEquivalent:@""];
	i.target = menu;
	i.tag = tag;
	i.state = on;
}

static OSStatus onHotkey(EventHandlerCallRef next, EventRef e, void *ctx) {
	EventHotKeyID hk;
	GetEventParameter(e, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof hk, NULL, &hk);
	goHotkey(hk.id == 2);
	return noErr;
}

void bzRun(bool paste, const void *icon, int iconLen, const char *installed) {
	[NSApplication sharedApplication];
	NSApp.activationPolicy = NSApplicationActivationPolicyAccessory;
	if (paste) // ask once for the permission bzPaste needs
		AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef) @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES});
	build();

	item = [NSStatusBar.systemStatusBar statusItemWithLength:NSSquareStatusItemLength];
	NSImage *dog = [[NSImage alloc] initWithData:[NSData dataWithBytes:icon length:iconLen]];
	dog.size = NSMakeSize(18, 18);
	dog.template = YES; // drawn in the menu bar's own colour
	item.button.image = dog;
	item.menu = [NSMenu new];
	item.menu.delegate = menu = [Menu new];

	// Ctrl+Alt+V steps to older clippings, with Shift to newer ones.
	EventTypeSpec pressed = {kEventClassKeyboard, kEventHotKeyPressed};
	InstallApplicationEventHandler(&onHotkey, 1, &pressed, NULL, NULL);
	EventHotKeyRef ref;
	RegisterEventHotKey(kVK_ANSI_V, controlKey | optionKey, (EventHotKeyID){'BLNC', 1}, GetApplicationEventTarget(), 0, &ref);
	RegisterEventHotKey(kVK_ANSI_V, controlKey | optionKey | shiftKey, (EventHotKeyID){'BLNC', 2}, GetApplicationEventTarget(), 0, &ref);

	// A poll: the pasteboard has no change notification, only a counter.
	__block NSInteger seen = NSPasteboard.generalPasteboard.changeCount;
	[NSTimer scheduledTimerWithTimeInterval:0.5 repeats:YES block:^(NSTimer *t) {
		NSPasteboard *pb = NSPasteboard.generalPasteboard;
		if (pb.changeCount == seen) return;
		seen = pb.changeCount;
		NSString *s = [pb stringForType:NSPasteboardTypeString];
		BOOL transient = hasType(@[@"de.petermaurer.TransientPasteboardType", @"com.typeit4me.clipping",
			@"Pasteboard generator type", @"org.nspasteboard.TransientType", @"org.nspasteboard.AutoGeneratedType"]);
		if (s && !transient) goClip((char *)s.UTF8String, hasType(@[@"com.apple.is-remote-clipboard"]));
	}];
	if (installed) { // said once the menu bar item is there to point at
		NSString *head = @(installed);
		dispatch_async(dispatch_get_main_queue(), ^{
			NSAlert *a = [NSAlert new];
			a.messageText = head ?: @"";
			a.informativeText = @"It is in the menu bar: look for the dog.";
			[NSApp activateIgnoringOtherApps:YES];
			[a runModal];
		});
	}
	[NSApp run];
}
