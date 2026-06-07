// Позитив: util не разрешён в pkg кастомного дерева — контроль
// работает на любом уровне, не только в internal/.
package util // want `GID-158: folder "util" is not allowed in pkg/ \(allowed: api, contract\); configure the tree via settings\.tree`

func Helper() {}
