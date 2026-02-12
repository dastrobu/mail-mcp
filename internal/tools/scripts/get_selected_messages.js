#!/usr/bin/osascript -l JavaScript

function run(argv) {
    const Mail = Application('Mail');
    Mail.includeStandardAdditions = true;
    
    // Parse arguments: limit
    const limit = parseInt(argv[0]);
    
    if (!limit || limit < 1) {
        return JSON.stringify({
            success: false,
            error: 'Limit is required and must be at least 1'
        });
    }
    
    if (limit > 100) {
        return JSON.stringify({
            success: false,
            error: 'Limit cannot exceed 100'
        });
    }
    
    try {
        // Get the selected messages from the frontmost Mail viewer
        const viewers = Mail.messageViewers();
        
        if (!viewers || viewers.length === 0) {
            return JSON.stringify({
                success: false,
                error: 'No Mail viewer windows are open'
            });
        }
        
        // Get the frontmost viewer
        const viewer = viewers[0];
        const selectedMessages = viewer.selectedMessages();
        
        if (!selectedMessages || selectedMessages.length === 0) {
            return JSON.stringify({
                success: true,
                data: {
                    count: 0,
                    messages: []
                }
            });
        }
        
        // Extract message details (limited by limit parameter)
        const result = [];
        const messagesToProcess = Math.min(selectedMessages.length, limit);
        for (let i = 0; i < messagesToProcess; i++) {
            const msg = selectedMessages[i];
            
            // Get mailbox and account information
            const mailbox = msg.mailbox();
            const account = mailbox.account();
            
            result.push({
                id: msg.id(),
                subject: msg.subject(),
                sender: msg.sender(),
                dateReceived: msg.dateReceived().toISOString(),
                dateSent: msg.dateSent().toISOString(),
                readStatus: msg.readStatus(),
                flaggedStatus: msg.flaggedStatus(),
                junkMailStatus: msg.junkMailStatus(),
                mailbox: mailbox.name(),
                account: account.name()
            });
        }
        
        return JSON.stringify({
            success: true,
            data: {
                count: result.length,
                messages: result
            }
        });
    } catch (e) {
        return JSON.stringify({
            success: false,
            error: e.toString()
        });
    }
}