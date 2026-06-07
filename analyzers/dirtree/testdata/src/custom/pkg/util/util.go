// Позитив: util не разрешён в pkg кастомного дерева — контроль
// работает на любом уровне, не только в internal/.
package util // want `GID-158: папка "util" не разрешена в pkg/ \(разрешены: api, contract\); дерево настраивается через settings\.tree`

func Helper() {}
