
dep:
	@$(MAKE) -C client dep
	@$(MAKE) -C discovery dep
	@$(MAKE) -C traceability dep


dep-branch:
	@$(MAKE) -C client dep-branch
	@$(MAKE) -C discovery dep-branch
	@$(MAKE) -C traceability dep-branch

dep-version:
	@$(MAKE) -C client dep-version sdk=$(sdk)
	@$(MAKE) -C discovery dep-version sdk=$(sdk)
	@$(MAKE) -C traceability dep-version sdk=$(sdk)

dep-sdk: 
	@$(MAKE) -C client dep-sdk
	@$(MAKE) -C discovery dep-sdk
	@$(MAKE) -C traceability dep-sdk
