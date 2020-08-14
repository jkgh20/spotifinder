import { shallowMount } from '@vue/test-utils'
import Selector from '../src/components/Selector.vue'

describe('Selector', () => {
    const wrapper = shallowMount(Selector, {
        propsData: {
            selectorName: "testSelector",
            selectedItems: ['Sin City', 'Atlantis'],
            availableItems: ['New Delhi', 'Loverstown', 'Penicillin'],
            maxItems: "4"
        }
    })

    it('populates selected items properly', () => {
        expect(wrapper.vm.selectedItems.length).toBe(2)

        expect(wrapper.text()).toContain('Sin City')
        expect(wrapper.text()).toContain('Atlantis')
        expect(wrapper.text()).not.toContain('New Delhi')
    })

    it('populates available items properly', () => {
        expect(wrapper.vm.availableItems.length).toBe(3)
    })

    it('toggles showSearchResults when search bar is focused or blurred', () => {
        const searchBar = wrapper.find('.searchBar')

        expect(wrapper.vm.showSearchResults).toBe(false)

        searchBar.trigger('focus')
        
        expect(wrapper.vm.showSearchResults).toBe(true)
        const availableItem = wrapper.findAll('.searchResultItem')

        searchBar.trigger('blur')
        expect(wrapper.vm.showSearchResults).toBe(false)   
    })

    it('filters available items properly', () => {
        const searchBar = wrapper.find('.searchBar')
        searchBar.trigger('focus')

        searchBar.element.value = 'lover'
        searchBar.trigger('input')
        expect(wrapper.vm.filteredItems.length).toBe(1)

        searchBar.element.value = 'i'
        searchBar.trigger('input')
        expect(wrapper.vm.filteredItems.length).toBe(2)
    })

    it('transfers values from selected items to available items', () => {
        const selectedItem = wrapper.find('.selectedItem')
        selectedItem.trigger('click')
        expect(wrapper.vm.selectedItems.length).toBe(1)
        expect(wrapper.vm.availableItems.length).toBe(4)
    })

    it('transfers values from available items to selected items', () => {
        const searchBar = wrapper.find('.searchBar')
        searchBar.trigger('focus')

        searchBar.element.value = 'lover'
        searchBar.trigger('input')

        expect(wrapper.vm.availableItems.length).toBe(4)

        const availableItem = wrapper.find('.searchResultItem')
        availableItem.trigger('mousedown')

        expect(wrapper.vm.availableItems.length).toBe(3)
    })

    it('prevents adding more selectedItems than the max', () => {
        /*
        const searchBar = wrapper.find('.searchBar')
        searchBar.trigger('focus')
        searchBar.element.value = ''
        searchBar.trigger('input')
        
        expect(wrapper.vm.selectedItems.length).toBe(2)
        expect(wrapper.vm.filteredItems.length).toBe(3)

        const availableItems = wrapper.findAll('.searchResultItem')

        availableItems.at(0).trigger('mousedown')
        expect(wrapper.vm.selectedItems.length).toBe(3)
        expect(wrapper.vm.filteredItems.length).toBe(2)

        availableItems.at(2).trigger('mousedown')
        expect(wrapper.vm.selectedItems.length).toBe(4)
        expect(wrapper.vm.filteredItems.length).toBe(1)

        availableItems.at(3).trigger('mousedown')
        expect(wrapper.vm.selectedItems.length).toBe(4)
        expect(wrapper.vm.filteredItems.length).toBe(1)
        */
    })
})