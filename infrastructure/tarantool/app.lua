box.once("bootstrap", function()
    box.schema.space.create('dialogs', { if_not_exists = true })
    box.space.dialogs:format({
        { name = 'id', type = 'unsigned' },
        { name = 'from', type = 'unsigned' },
        { name = 'to', type = 'unsigned' },
        { name = 'message', type = 'string' },
        { name = 'created_at', type = 'string' },
    })
    box.space.dialogs:create_index('id', { type = 'TREE', parts = { {'id', 'unsigned'} } })
    box.space.dialogs:create_index('from_to', { type = 'TREE', unique = false, parts = { { 'from', 'unsigned' }, { 'to', 'unsigned' } } })
end)
